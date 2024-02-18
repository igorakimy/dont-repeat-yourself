package cmd

import (
	"fmt"
	"github.com/igorakimy/dont-repeat-yourself/database"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	balancesCmd.AddCommand(balancesListCmd)
	dryCmd.AddCommand(balancesCmd)
}

var balancesCmd = &cobra.Command{
	Use:   "balances",
	Short: "Взаимодействие с балансами (list...).",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return incorrectUsageErr()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Список всех балансов.",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := database.NewStateFromDisk()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer state.Close()

		fmt.Println("Балансы на аккаунтах")
		fmt.Println("__________________")
		for account, balance := range state.Balances {
			fmt.Println(fmt.Sprintf("%s: %d", account, balance))
		}
	},
}

func incorrectUsageErr() error {
	return nil
}
