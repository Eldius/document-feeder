package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nmyk.io/cowsay"
)

// askCmd represents the ask command.
var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask a question to the model using the stored content",
	Long:  `Ask a question to the model using the stored content.`,
	Run: func(cmd *cobra.Command, args []string) {
		cancel := ui.ProcessingScreen(cmd.Context(), "Processing questionOut...")
		defer cancel()
		start := time.Now()
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			panic(err)
		}
		questionIn := strings.Join(args, " ")
		answer, err := a.AskAQuestion(cmd.Context(), questionIn)
		if err != nil {
			panic(err)
		}

		cancel()

		time.Sleep(1 * time.Second)

		questionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		answerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(false)

		footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("123")).
			Italic(true).
			Bold(true)

		fmt.Println("")
		fmt.Println("")
		fmt.Println("---")
		questionOut := "Question: " + questionIn
		fmt.Println(questionStyle.Render(questionOut))

		answerOut := "Answer: " + answer
		if askOpts.cowSay {
			cowsay.Cowsay(answerStyle.Render(answerOut))
		} else {
			fmt.Println(answerStyle.Render(answerOut))
		}

		if askOpts.outputFile != "" {
			fmt.Println("Outputting to file:", askOpts.outputFile)
			err := os.WriteFile(askOpts.outputFile, []byte(questionOut+"\n"+answerOut), 0644)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}
		}

		fmt.Println("")
		fmt.Println("")
		fmt.Println("")
		fmt.Println("---")
		fmt.Println("---")
		fmt.Println(footerStyle.Render(fmt.Sprintf("Time elapsed: %s", time.Since(start).String())))
	},
}

var (
	askOpts struct {
		cowSay       bool
		outputFile   string
		disableCache bool
	}
)

func init() {
	rootCmd.AddCommand(askCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// askCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// askCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	askCmd.Flags().BoolVarP(&askOpts.cowSay, "cow-say", "c", false, "Use cowsay to render the answer")
	askCmd.Flags().StringVarP(&askOpts.outputFile, "output-file", "o", "", "Output the answer to a file")
	askCmd.Flags().BoolVarP(&askOpts.disableCache, "no-cache", "d", false, "Disable caching for this question")
	if err := viper.BindPFlag("ollama.generation.no-cache", askCmd.Flags().Lookup("no-cache")); err != nil {
		fmt.Println("Failed to bind property:", err)
	}
}
