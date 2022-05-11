/*
Copyright © 2022 Amaan Mohib [amaan.mohib@gmail.com]

*/
package main

import (
	"cli/cmd"
	"cli/internal"
	"cli/utils"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		pathArr := strings.Split(path, "/")
		os.MkdirAll(strings.Join(pathArr[:len(pathArr)-1], "/"), 0755)

		os.Create(path)
		return true, nil
	}
	return false, err
}

func initConfig(path string) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.ReadInConfig()
}

func main() {
	home, err := os.UserConfigDir()
	utils.CheckError(err)

	configExists, configErr := exists(home + "/.urldump/config.json")
	if configExists && configErr == nil {
		initConfig(home + "/.urldump")
	} else {
		utils.CheckError(configErr)
	}

	dbPath := home + "/.urldump/urldump.sqlite"
	dbExists, dbErr := exists(dbPath)
	if dbExists && dbErr == nil {
		db := internal.SQLConnInit(dbPath)
		internal.SetSQLConn(db.DB)

	} else {
		utils.CheckError(dbErr)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading the .env file: %v", err)
	}

	cmd.Execute()
}
