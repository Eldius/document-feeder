package cmd

import (
	"github.com/eldius/document-feed-embedder/internal/adapter"
	"github.com/spf13/cobra"
)

// feedListCmd represents the list command
var feedListCmd = &cobra.Command{
	Use:   "list",
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
		feeds, err := a.All(cmd.Context())
		if err != nil {
			panic(err)
		}
		for _, f := range feeds {
			println(f.Title)
		}
	},
}

func init() {
	feedCmd.AddCommand(feedListCmd)
}
