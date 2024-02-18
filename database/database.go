package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Snapshot является снимком базы данных после совершения каждой транзакции
type Snapshot [32]byte

// Account является клиентом в базе данных
type Account string

func NewAccount(name string) Account {
	return Account(name)
}

// Tx отражает изменение в базе данных
type Tx struct {
	From  Account `json:"from"`
	To    Account `json:"to"`
	Value uint    `json:"value"`
	Data  string  `json:"data"`
}

func NewTx(from, to Account, value uint, data string) Tx {
	return Tx{
		from,
		to,
		value,
		data,
	}
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

	dbFile   *os.File
	snapshot Snapshot
}

func NewStateFromDisk() (*State, error) {
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
		Snapshot{},
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
func (s *State) Persist() (Snapshot, error) {
	// Создаем копию memPool, потому что s.txMemPool будет
	// изменен в цикле ниже
	memPool := make([]Tx, len(s.txMemPool))
	copy(memPool, s.txMemPool)

	for i := 0; i < len(memPool); i++ {
		txJson, err := json.Marshal(memPool[i])
		if err != nil {
			return Snapshot{}, err
		}

		fmt.Printf("Сохранение новой транзакции(TX) на диск:\n")
		fmt.Printf("\t%s\n", txJson)
		txJson = append(txJson, '\n')
		if _, err = s.dbFile.Write(txJson); err != nil {
			return Snapshot{}, err
		}

		if err = s.doSnapshot(); err != nil {
			return Snapshot{}, nil
		}
		fmt.Printf("Новый снимок БД: %x\n", s.snapshot)

		// Удалить TX, записанный в файл из memPool
		s.txMemPool = s.txMemPool[1:]
	}

	return s.snapshot, nil
}

// apply изменение и валидация состояния (пользовательских балансов)
func (s *State) apply(tx Tx) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf(
			"Недостаточно средств на балансе: нужно %d, есть %d\n",
			tx.Value,
			s.Balances[tx.From],
		)
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

func (s *State) doSnapshot() error {
	// Перечитать весь файл начиная с первого байта
	_, err := s.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}

	txsData, err := io.ReadAll(s.dbFile)
	if err != nil {
		return err
	}
	s.snapshot = sha256.Sum256(txsData)

	return nil
}

func (s *State) Close() {
	_ = s.dbFile.Close()
}
