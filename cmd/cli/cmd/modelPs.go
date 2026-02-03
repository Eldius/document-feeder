package cmd

import (
	"fmt"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/document-feeder/internal/ui"

	"github.com/spf13/cobra"
)

// modelPsCmd represents the ps command.
var modelPsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running models.",
	Long:  `List running models.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		c := ollama.NewOllamaClient()
		models, err := c.RunningModels(cmd.Context())
		if err != nil {
			err := fmt.Errorf("listing models: %w", err)
			fmt.Printf("failed to list models: %s\n", err)
			return err
		}

		ui.DisplayModels(models)

		return nil
	},
}

func init() {
	modelsCmd.AddCommand(modelPsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modelPsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// modelPsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
