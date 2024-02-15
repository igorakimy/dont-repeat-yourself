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
	Short: "Interact with balances (list...).",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return incorrectUsageErr()
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all balances.",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := database.NewStateFromDisk()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer state.Close()

		fmt.Println("Accounts balances:")
		fmt.Println("__________________")
		for account, balance := range state.Balances {
			fmt.Println(fmt.Sprintf("%s: %d", account, balance))
		}
	},
}

func incorrectUsageErr() error {
	return nil
}
