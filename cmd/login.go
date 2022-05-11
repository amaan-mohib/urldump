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

func auth(clientID, audience, clientSecret, authDomain string) {
	internal.AuthorizeUser(clientID, audience, clientSecret, authDomain, "http://127.0.0.1:14565/oauth/callback/")
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
		var clientID = os.Getenv("AUTH0_CLIENT_ID")
		var authDomain = os.Getenv("AUTH0_DOMAIN")
		var audience = os.Getenv("AUTH0_AUDIENCE")
		var clientSecret = os.Getenv("AUTH0_CLIENT_SECRET")

		if reauth {
			auth(clientID, audience, clientSecret, authDomain)
			os.Exit(0)
		}
		accessToken := viper.Get("tokens.AccessToken")
		expiresAt := viper.Get("tokens.ExpiresAt")

		if expiresAt != nil && expiresAt.(float64) < float64(time.Now().Unix()) {
			internal.RefreshToken(clientID, clientSecret)
			name := viper.Get("user.name")
			email := viper.Get("user.email")
			fmt.Printf("\nAlready logged in as %v [%v]\n", name, email)
		} else {
			if accessToken != nil {
				isValid, err := internal.ValidateToken(accessToken.(string))
				if !isValid && err != nil {
					auth(clientID, audience, clientSecret, authDomain)
				} else {
					name := viper.Get("user.name")
					email := viper.Get("user.email")
					fmt.Printf("\nAlready logged in as %v [%v]\n", name, email)
				}
			} else {
				auth(clientID, audience, clientSecret, authDomain)
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
