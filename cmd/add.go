/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"cli/internal"
	"cli/models"
	"cli/utils"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	URL   string
	Title string
	Tags  []string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// url, errURL := cmd.Flags().GetString("url")
		// utils.CheckError(errURL)
		if URL != "" {
			db := internal.GetSQLConn().DB

			stmt, err := db.Prepare("INSERT INTO links (url, title, tags) VALUES (?, ?, ?)")
			utils.CheckError(err)

			link := models.Link{
				URL:   URL,
				Title: Title,
				Tags:  strings.Join(Tags, ","),
			}
			stmt.Exec(link.URL, link.Title, link.Tags)
			defer stmt.Close()

			fmt.Printf("\n✓ Item added\n----\nURL: %v\nTitle: %v\nTags: %v\n", link.URL, link.Title, link.Tags)
		} else {
			fmt.Println("please add url")
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	addCmd.Flags().StringVarP(&URL, "url", "u", "", "Add the given link")
	addCmd.Flags().StringVarP(&Title, "title", "t", "", "Add the title of given link")
	addCmd.Flags().StringSliceVarP(&Tags, "tags", "g", []string{""}, "Add tags for the given link (list or comma separated)")
}
