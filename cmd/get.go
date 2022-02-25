/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"cli/internal"
	"cli/models"
	"cli/utils"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var all bool
var list bool
var limit int
var order string
var status string

func makeList(rows *sql.Rows) []models.Link {
	links := make([]models.Link, 0)

	for rows.Next() {
		link := models.Link{}
		err := rows.Scan(&link.ID, &link.URL, &link.Title, &link.Tags, &link.Status)
		utils.CheckError(err)
		links = append(links, link)
	}
	err := rows.Err()
	utils.CheckError(err)

	return links
}

func getAll(db *sql.DB) []models.Link {
	query := func() string {
		if status == "all" {
			return fmt.Sprintf("SELECT * FROM links ORDER BY title %v", order)
		} else {
			return fmt.Sprintf("SELECT * FROM links WHERE status = \"%v\" ORDER BY title %v", status, order)
		}
	}
	rows, err := db.Query(query())
	utils.CheckError(err)
	defer rows.Close()

	err = rows.Err()
	utils.CheckError(err)

	links := makeList(rows)

	return links
}

func formatRes(links []models.Link) {
	fmt.Printf("Found %v links\n", len(links))

	for _, link := range links {
		fmt.Printf("\n----\nID: %v\nURL: %v\nTitle: %v\nTags: %v\nStatus: %v\n", link.ID, link.URL, link.Title, link.Tags, strings.Title(link.Status))
	}
}

func listLinks(db *sql.DB) []models.Link {
	query := func() string {
		if status == "all" {
			return fmt.Sprintf("SELECT * FROM links ORDER BY title %v LIMIT %v", order, limit)
		} else {
			return fmt.Sprintf("SELECT * FROM links WHERE status = \"%v\" ORDER BY title %v LIMIT %v", status, order, limit)
		}
	}
	rows, err := db.Query(query())
	utils.CheckError(err)
	defer rows.Close()

	err = rows.Err()
	utils.CheckError(err)

	links := makeList(rows)

	return links
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "This command will fetch the article with the given ID",
	Long:  `This command will fetch the article or website with the given ID`,
	Run: func(cmd *cobra.Command, args []string) {
		db := internal.GetSQLConn().DB

		if order != "asc" && order != "desc" {
			log.Fatalln("Invalid order value")
		}
		if status != "all" && status != "read" && status != "unread" {
			log.Fatalln("Invalid status value")
		}
		if all && list {
			log.Fatalln("Please do not use all and list together")
		}
		if all || len(os.Args) == 2 {
			links := getAll(db)
			formatRes(links)
		}
		if list {
			links := listLinks(db)
			fmt.Printf("\nList of limit %v and ordered by %v\n", limit, order)
			formatRes(links)
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	getCmd.Flags().BoolVarP(&all, "all", "a", false, "Get all links")
	getCmd.Flags().BoolVarP(&list, "list", "f", false, "Get list of limited links (default 10)")
	getCmd.Flags().IntVarP(&limit, "limit", "l", 10, "Set limit to show links")
	getCmd.Flags().StringVarP(&order, "order", "o", "asc", "Set order for the fetched links [asc | desc]")
	getCmd.Flags().StringVarP(&status, "status", "s", "all", "Get links with the given status [read | unread | all]")
}
