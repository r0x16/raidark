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

func main() {
	err := godotenv.Load(".env")

	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cmd.Execute()
}
