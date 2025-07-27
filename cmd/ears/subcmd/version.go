// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package subcmd

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
