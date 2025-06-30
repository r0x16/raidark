/*
Copyright © 2024 r0x16
*/
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/r0x16/Raidark/cmd"
)

// loadEnvIfExists loads the .env file only if it exists
func loadEnvIfExists() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
		log.Println("Environment file loaded successfully")
	}
}

func main() {
	loadEnvIfExists()
	cmd.Execute()
}
