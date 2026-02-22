package cmd

import (
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display benchmark results for a given models",
	Long:  `Display benchmark results for a given models.`,
	Run: func(cmd *cobra.Command, args []string) {
		a, err := adapter.NewBenchmarkFromConfigs()
		if err != nil {
			panic(err)
		}
		if err := a.Plot(cmd.Context(), showOpts.models); err != nil {
			panic(err)
		}

		fmt.Println("done")
	},
}

var (
	showOpts struct {
		models []string
	}
)

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.PersistentFlags().StringSliceVar(&showOpts.models, "model", []string{"deepseek-r1:1.5b", "llama3:8b-instruct-q4_K_M", "tinyllama:latest"}, "ollama model to use")
}
