package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password_plain := "adminpassword"

	// Hash admin password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password_plain), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error hashing admin password: %v", err)
	}
	fmt.Printf("Hashed Password for '%s': %s\n", password_plain, string(hashed))

	fmt.Println("\nCopy these hashed passwords and replace the placeholders in your SQL INSERT statements.")
}
