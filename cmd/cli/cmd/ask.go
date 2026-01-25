package cmd

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/adapter"
	"github.com/spf13/cobra"
	"nmyk.io/cowsay"
	"strings"
	"time"
)

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask a question to the model using the stored content",
	Long:  `Ask a question to the model using the stored content.`,
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			panic(err)
		}
		answer, err := a.AskAQuestion(cmd.Context(), strings.Join(args, " "))
		if err != nil {
			panic(err)
		}
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
		fmt.Println()
		fmt.Println()
		fmt.Println("---")
		fmt.Println(questionStyle.Render("Question: " + strings.Join(args, " ")))

		if askOpts.cowSay {
			cowsay.Cowsay(answerStyle.Render("Answer: " + answer))
		} else {
			fmt.Println(answerStyle.Render("Answer: " + answer))
		}

		fmt.Println()
		fmt.Println()
		fmt.Println()
		fmt.Println("---")
		fmt.Println("---")
		fmt.Println(footerStyle.Render(fmt.Sprintf("Time elapsed: %s", time.Since(start).String())))
	},
}

var (
	askOpts struct {
		cowSay bool
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
}
