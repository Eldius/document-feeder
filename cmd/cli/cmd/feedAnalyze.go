package cmd

import (
	"fmt"
	"strings"

	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/ui/v2/analyze_feed"
	"github.com/spf13/cobra"
)

// feedAnalyzeCmd represents the analyze command.
var feedAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze feed articles for subjects, summaries, and sentiment",
	Long: `Analyze feed articles for subjects, summaries, and sentiment.
Example: feed analyze --link https://example.com/rss.xml
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := adapter.NewFeedAdapterFromConfigs()
		if err != nil {
			return fmt.Errorf("creating adapter: %w", err)
		}

		if feedAnalyzeOpts.interactive {
			return analyze_feed.Start(cmd.Context(), a, feedAnalyzeOpts.link, feedAnalyzeOpts.noCache)
		}

		results, err := a.AnalyzeFeed(cmd.Context(), feedAnalyzeOpts.link, feedAnalyzeOpts.noCache)
		if err != nil {
			return fmt.Errorf("analyzing feed: %w", err)
		}

		for link, res := range results {
			fmt.Printf("\n--- Article: %s ---\n", link)
			fmt.Printf("Subject:   %s\n", res.Subject)
			fmt.Printf("Sentiment: %s\n", res.Sentiment)
			fmt.Printf("Summary:   %s\n", res.Summary)
			fmt.Printf("Keywords:  %s\n", strings.Join(res.Keywords, ", "))
		}

		return nil
	},
}

var (
	feedAnalyzeOpts struct {
		link        string
		noCache     bool
		interactive bool
	}
)

func init() {
	feedCmd.AddCommand(feedAnalyzeCmd)

	feedAnalyzeCmd.Flags().StringVarP(&feedAnalyzeOpts.link, "link", "l", "", "feed link to analyze")
	feedAnalyzeCmd.Flags().BoolVarP(&feedAnalyzeOpts.noCache, "no-cache", "n", false, "force re-analysis by skipping cache")
	feedAnalyzeCmd.Flags().BoolVarP(&feedAnalyzeOpts.interactive, "interactive", "i", false, "Visual execution feedback for the user.")
	_ = feedAnalyzeCmd.MarkFlagRequired("link")
}
