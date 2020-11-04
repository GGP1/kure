package main

import (
	"path/filepath"

	// jpeg and png imported for displaying images on the terminal
	_ "image/jpeg"
	_ "image/png"
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

	path := filepath.Join(dbPath, dbName)

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("couldn't find user home directory", err)
		}
		path = filepath.Join(home, "kure")
	}

	if err := db.Init(path); err != nil {
		log.Fatal("database error: ", err)
	}

	cmd.Execute()
}
