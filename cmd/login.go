/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"cli/internal"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reauth bool

func auth() {
	internal.AuthorizeUser("MaMKxlV3NAhQJPOjeBIr5dKRb21VQsaL", "https://dev-y3bdlwyd.us.auth0.com/api/v2/", "G2e-gAOgyQv4Y_3kHW9p7x7XxKTFZPh9zXMdUGPlDHfEPaqTlTYWB_18Q9GOPXG6", "dev-y3bdlwyd.us.auth0.com", "http://127.0.0.1:14565/oauth/callback/")
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if reauth {
			auth()
			os.Exit(0)
		}
		accessToken := viper.Get("tokens.AccessToken")
		expiresAt := viper.Get("tokens.ExpiresAt")
		// fmt.Println(accessToken)
		if expiresAt != nil && expiresAt.(float64) < float64(time.Now().Unix()) {
			internal.RefreshToken("MaMKxlV3NAhQJPOjeBIr5dKRb21VQsaL", "G2e-gAOgyQv4Y_3kHW9p7x7XxKTFZPh9zXMdUGPlDHfEPaqTlTYWB_18Q9GOPXG6")
			name := viper.Get("user.name")
			email := viper.Get("user.email")
			fmt.Printf("\nAlready logged in as %v [%v]\n", name, email)
		} else {
			if accessToken != nil {
				isValid, err := internal.ValidateToken(accessToken.(string))
				if !isValid && err != nil {
					auth()
				} else {
					name := viper.Get("user.name")
					email := viper.Get("user.email")
					fmt.Printf("\nAlready logged in as %v [%v]\n", name, email)
				}
			} else {
				auth()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	loginCmd.Flags().BoolVarP(&reauth, "reauth", "r", false, "Reauthorize")
}
