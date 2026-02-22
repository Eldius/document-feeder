/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/eldius/document-feeder/internal/ui"

	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/model"
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
		a, err := adapter.NewFeedAdapterFromConfigs()
		if err != nil {
			err := fmt.Errorf("creating adapter: %w", err)
			fmt.Printf("failed to create adapter: %s\n", err)
			return err
		}
		fmt.Println("searching feeds")
		var res []*model.SearchResult
		question := strings.Join(args, " ")
		if feedSearchOpts.similarityThreshold == 0 {
			res, err = a.Search(cmd.Context(), question, feedSearchOpts.maxResults)
			if err != nil {
				err := fmt.Errorf("searching feeds: %w", err)
				fmt.Printf("failed to search feeds: %s\n", err)
				return err
			}
		} else {
			res, err = a.SearchWithSimilarityThreshold(cmd.Context(), question, feedSearchOpts.maxResults, feedSearchOpts.similarityThreshold)
			if err != nil {
				err := fmt.Errorf("searching feeds: %w", err)
				fmt.Printf("failed to search feeds: %s\n", err)
				return err
			}
		}

		if err := ui.ContentReader(cmd.Context(), res, question); err != nil {
			fmt.Printf("failed to read articles: %s\n", err)
			return err
		}

		return nil
	},
}

var (
	feedSearchOpts struct {
		maxResults          int
		similarityThreshold float32
	}
)

func init() {
	feedCmd.AddCommand(feedSearchCmd)

	feedSearchCmd.Flags().IntVarP(&feedSearchOpts.maxResults, "max-results", "m", 10, "max results to return")
	feedSearchCmd.Flags().Float32VarP(&feedSearchOpts.similarityThreshold, "similarity-threshold", "s", 0, "similarity threshold")
}
