package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const Major = "0"
const Minor = "1"
const Fix = "0"
const Verbal = "Добавление транзакций && Список балансов"

func init() {
	dryCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Описание версии.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Версия: %s.%s.%s-beta %s\n", Major, Minor, Fix, Verbal)
	},
}
