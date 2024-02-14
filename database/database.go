package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Account является клиентом в базе данных
type Account string

// Tx отражает изменение в базе данных
type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

// IsReward проверяет, являются ли данные "вознаграждением",
// которое необходимо для искусственного поддержания токена
// (процесс инфляции)
func (t Tx) IsReward() bool {
	return t.Data == "reward"
}

// State является самым важным компонентом базы данных,
// который инкапсулирует всю бизнес-логику и знает о всех балансах
// клиентов, о том, кто передавал токены и сколько их было передано
type State struct {
	Balances  map[Account]uint
	txMemPool []Tx

	dbFile *os.File
}

func NewStateFromDist() (*State, error) {
	// получить текущую директорию
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	genFilePath := filepath.Join(cwd, "database", "genesis.json")
	gen, err := loadGenesis(genFilePath)
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	txDbFilePath := filepath.Join(cwd, "database", "tx.db")
	dbFile, err := os.OpenFile(txDbFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(dbFile)
	state := &State{
		balances,
		make([]Tx, 0),
		dbFile,
	}

	// Итерировать по каждой строке файла tx.db
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		// Конвертировать закодированный TX из JSON в структуру
		var tx Tx
		_ = json.Unmarshal(scanner.Bytes(), &tx)

		// Восстановить состояние (пользовательские балансы),
		// как серию событий
		if err := state.apply(tx); err != nil {
			return nil, err
		}
	}

	return state, nil
}

// Add добавление новой транзакции в memPool
func (s *State) Add(tx Tx) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMemPool = append(s.txMemPool, tx)

	return nil
}

// Persist сохранение транзакций на диск
func (s *State) Persist() error {
	// Создаем копию memPool, потому что s.txMemPool будет
	// изменен в цикле ниже
	memPool := make([]Tx, len(s.txMemPool))
	copy(memPool, s.txMemPool)

	for i := 0; i < len(memPool); i++ {
		txJson, err := json.Marshal(memPool[i])
		if err != nil {
			return err
		}

		txJson = append(txJson, '\n')
		if _, err = s.dbFile.Write(txJson); err != nil {
			return err
		}

		// Удалить TX, записанный в файл из memPool
		s.txMemPool = s.txMemPool[1:]
	}

	return nil
}

// apply изменение и валидация состояния (пользовательских балансов)
func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("insufficient balance")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

func loadGenesis(filePath string) (*State, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	var state *State
	_ = json.NewDecoder(f).Decode(state)

	return state, nil
}
