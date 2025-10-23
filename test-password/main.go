package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Ambil HASH dari log Anda
	hashFromDB := "$2a$04$WL6VfZpzqEaJ2LtyPjl1ceKoP1abKXKP4y3NyRgADJc8snLajuA8"

	// Ambil PASSWORD dari log Anda
	passwordFromLogin := "keputih123"

	fmt.Printf("Membandingkan...\n")
	fmt.Printf("Hash:     %s\n", hashFromDB)
	fmt.Printf("Password: %s\n", passwordFromLogin)

	// Ini adalah fungsi yang sama dengan helpers.CheckPassword
	err := bcrypt.CompareHashAndPassword([]byte(hashFromDB), []byte(passwordFromLogin))

	if err != nil {
		log.Fatalf("Hasil: Password SALAH. (Error: %v)", err)
	}

	fmt.Println("Hasil: Password BENAR.")
}
