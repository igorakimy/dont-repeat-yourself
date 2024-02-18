package cmd

import (
	"fmt"
	"github.com/igorakimy/dont-repeat-yourself/database"
	"github.com/spf13/cobra"
	"os"
)

const (
	flagFrom  = "from"
	flagTo    = "to"
	flagValue = "value"
	flagData  = "data"
)

func init() {
	txAddCmd.Flags().String(flagFrom, "", "С какого аккаунта отправить токены")
	_ = txAddCmd.MarkFlagRequired(flagFrom)

	txAddCmd.Flags().String(flagTo, "", "На какой аккаунт отправить токены")
	_ = txAddCmd.MarkFlagRequired(flagTo)

	txAddCmd.Flags().Uint(flagValue, 0, "Сколько токенов отправить")
	_ = txAddCmd.MarkFlagRequired(flagValue)

	txAddCmd.Flags().String(flagData, "", "Допустимые значения: 'reward'")

	txsCmd.AddCommand(txAddCmd)

	dryCmd.AddCommand(txsCmd)
}

var txsCmd = &cobra.Command{
	Use:   "tx",
	Short: "Взаимодействие с транзакциями (add...)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return incorrectUsageErr()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var txAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Добавление новой транзакции(TX) в базу данных.",
	Run: func(cmd *cobra.Command, args []string) {
		from, _ := cmd.Flags().GetString(flagFrom)
		to, _ := cmd.Flags().GetString(flagTo)
		value, _ := cmd.Flags().GetUint(flagValue)
		data, _ := cmd.Flags().GetString(flagData)

		fromAcc := database.NewAccount(from)
		toAcc := database.NewAccount(to)

		tx := database.NewTx(fromAcc, toAcc, value, data)

		state, err := database.NewStateFromDisk()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Закрываем файл бд со всеми транзакциями
		defer state.Close()

		// Добавляем транзакцию(TX) в массив(пул) в памяти
		if err := state.Add(tx); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Сбросить данные из пула в памяти на диск
		_, err = state.Persist()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("Транзакция(TX) успешно добавлена в реестр.")
	},
}
