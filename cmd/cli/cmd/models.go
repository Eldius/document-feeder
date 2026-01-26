package cmd

import (
	"fmt"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/spf13/cobra"
)

// modelsCmd represents the models command.
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models.",
	Long:  `List all available models.`,
	Run: func(cmd *cobra.Command, args []string) {
		c := ollama.NewOllamaClient()
		models, err := c.ListModels(cmd.Context())
		if err != nil {
			panic(err)
		}
		for _, m := range models.Models {
			fmt.Println("")
			fmt.Println("- name:    ", m.Name)
			fmt.Println("  size:    ", m.Size)
			fmt.Println("  modified:", m.ModifiedAt.Format("2006-01-02 15:04:05"))
			fmt.Println("  details:")
			fmt.Println("     format:", m.Details.Format)
			fmt.Println("     families:")
			for _, f := range m.Details.Families {
				fmt.Println("        -", f)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modelsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// modelsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
