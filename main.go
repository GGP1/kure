package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/GGP1/kure/cmd"
	"github.com/GGP1/kure/db"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Create database in the user home directory
	if err := db.Init(home); err != nil {
		log.Fatal("error: ", err)
	}

	cmd.RootCmd.Execute()
}
