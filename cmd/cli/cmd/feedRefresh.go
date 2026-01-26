package cmd

import (
	"fmt"

	"github.com/eldius/document-feeder/internal/adapter"

	"github.com/spf13/cobra"
)

// feedRefreshCmd represents the refresh command.
var feedRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			panic(err)
		}
		fmt.Println("refreshing feeds")
		if err := a.Refresh(cmd.Context()); err != nil {
			panic(err)
		}
	},
}

func init() {
	feedCmd.AddCommand(feedRefreshCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
