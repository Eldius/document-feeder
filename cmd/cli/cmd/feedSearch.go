/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"
	"strings"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var feedSearchCmd = &cobra.Command{
	Use:   "search",
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
		fmt.Println("searching feeds")
		articles, err := a.Search(cmd.Context(), strings.Join(args, " "), feedSearchOpts.maxResults)
		if err != nil {
			panic(err)
		}
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		feedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Bold(false)
		tagStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("123")).
			Bold(true)
		fmt.Println()
		fmt.Println()
		fmt.Println("------")
		fmt.Println(titleStyle.Render("Found articles in feeds for '" + strings.Join(args, " ") + "':"))
		for _, a := range articles {
			fmt.Println(feedStyle.Render(" ->", a.Article.Title), tagStyle.Render(" => (", fmt.Sprintf("%0.2f", a.Similarity), strings.Join(a.Article.Categories, ", "), ")"))
		}
	},
}

var (
	feedSearchOpts struct {
		maxResults int
	}
)

func init() {
	feedCmd.AddCommand(feedSearchCmd)

	feedSearchCmd.Flags().IntVarP(&feedSearchOpts.maxResults, "max-results", "m", 10, "max results to return")
}
