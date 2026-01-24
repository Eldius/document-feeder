/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/adapter"
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
		feeds, err := a.Search(cmd.Context(), strings.Join(args, " "))
		if err != nil {
			panic(err)
		}
		for _, f := range feeds {
			fmt.Println(f.Title)
		}
	},
}

func init() {
	feedCmd.AddCommand(feedSearchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
