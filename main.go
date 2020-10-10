package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db"

	"github.com/spf13/viper"
)

func main() {
	// Load config file and env variables
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	dbPath := strings.TrimSuffix(viper.GetString("database.path"), "/")
	dbName := viper.GetString("database.name")

	path := fmt.Sprintf("%s/%s.db", dbPath, dbName)

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("couldn't find user home directory", err)
		}
		path = fmt.Sprintf("%s/kure.db", home)
	}

	if err := db.Init(path); err != nil {
		log.Fatal("database error: ", err)
	}

	cmd.Execute()
}
