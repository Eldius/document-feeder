package cmd

import (
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/adapter"
	"github.com/spf13/cobra"
)

// feedAddCmd represents the add command
var feedAddCmd = &cobra.Command{
	Use:   "add",
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
		for _, f := range feedAddOpts.feed {
			feed, err := a.Parse(cmd.Context(), f)
			if err != nil {
				panic(err)
			}
			fmt.Println("==>", feed.Title, "<==")
		}
	},
}

var (
	feedAddOpts struct {
		feed []string
	}
)

func init() {
	feedCmd.AddCommand(feedAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// feedAddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// feedAddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	feedAddCmd.Flags().StringSliceVarP(&feedAddOpts.feed, "feed", "f", []string{}, "feed to add")
}
