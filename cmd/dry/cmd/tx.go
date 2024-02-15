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
	txAddCmd.Flags().String(flagFrom, "", "From what account to send tokens")
	_ = txAddCmd.MarkFlagRequired(flagFrom)

	txAddCmd.Flags().String(flagTo, "", "To what account to send tokens")
	_ = txAddCmd.MarkFlagRequired(flagTo)

	txAddCmd.Flags().Uint(flagValue, 0, "How many tokens to send")
	_ = txAddCmd.MarkFlagRequired(flagValue)

	txAddCmd.Flags().String(flagData, "", "Possible values: 'reward'")

	txsCmd.AddCommand(txAddCmd)

	dryCmd.AddCommand(txsCmd)
}

var txsCmd = &cobra.Command{
	Use:   "tx",
	Short: "Interact with txs (add...)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return incorrectUsageErr()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var txAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds new TX to database.",
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
		if err := state.Persist(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("TX successfully added to the ledger...")
	},
}
