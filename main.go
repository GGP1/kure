package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db"

	"github.com/spf13/viper"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load config file and env variables
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}

	dbPath := viper.GetString("database.path")
	dbPath = strings.TrimSuffix(dbPath, "/")

	dbName := viper.GetString("database.name")

	path := fmt.Sprintf("%s/%s.db", dbPath, dbName)

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		path = fmt.Sprintf("%s/kure.db", home)
	}

	// Create database in the path specified
	if err := db.Init(path); err != nil {
		log.Fatal("database error: ", err)
	}

	cmd.RootCmd.Execute()
}
