package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xmidt-org/ears/pkg/app"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the app version information",
	Long:  "Prints that app version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(app.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
