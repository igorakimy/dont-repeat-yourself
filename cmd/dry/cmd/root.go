package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var dryCmd = &cobra.Command{
	Use:   "dry",
	Short: "Don't Repeat Yourself CLI",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	if err := dryCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
