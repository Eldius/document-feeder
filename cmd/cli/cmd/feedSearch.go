/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/document-feeder/internal/ui"

	"strings"

	"github.com/spf13/cobra"
)

// feedSearchCmd represents the search command.
var feedSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := adapter.NewDefaultAdapter()
		if err != nil {
			err := fmt.Errorf("creating adapter: %w", err)
			fmt.Printf("failed to create adapter: %w", err)
			return err
		}
		fmt.Println("searching feeds")
		res, err := a.Search(cmd.Context(), strings.Join(args, " "), feedSearchOpts.maxResults)
		if err != nil {
			err := fmt.Errorf("searching feeds: %w", err)
			fmt.Printf("failed to search feeds: %w", err)
			return err
		}

		var articles []model.Article
		for _, a := range res {
			articles = append(articles, a.Article)
		}

		if err := ui.ArticleReaderScreen(cmd.Context(), articles); err != nil {
			err := fmt.Errorf("reading articles: %w", err)
			fmt.Printf("failed to read articles: %w", err)
			return err
		}
		return nil
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
