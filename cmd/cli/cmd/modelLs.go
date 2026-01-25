package cmd

import (
	"github.com/eldius/document-feed-embedder/internal/client/ollama"
	"github.com/eldius/document-feed-embedder/internal/ui"

	"github.com/spf13/cobra"
)

// modelLsCmd represents the ls command
var modelLsCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List available models.",
	Long:    `List available models.`,
	Run: func(cmd *cobra.Command, args []string) {

		c := ollama.NewOllamaClient()
		models, err := c.ListModels(cmd.Context())
		if err != nil {
			panic(err)
		}
		ui.DisplayModels(models)
	},
}

func init() {
	modelsCmd.AddCommand(modelLsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modelLsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// modelLsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
