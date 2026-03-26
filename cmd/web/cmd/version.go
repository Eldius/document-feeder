package cmd

import (
	"github.com/eldius/document-feeder/internal/config"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of the application",
	Long:  `Displays the version of the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf(`---
version: %v
commit: %v
build date: %v`, config.Version, config.Commit, config.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
