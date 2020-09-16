package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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

	path := viper.GetString("db_path")

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
