package cmd

import (
	"github.com/spf13/cobra"
)

// feedCmd represents the feed command.
var feedCmd = &cobra.Command{
	Use:   "feed",
	Short: "Feed related commands.",
	Long:  `Feed related commands.`,
}

func init() {
	rootCmd.AddCommand(feedCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// feedCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// feedCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
