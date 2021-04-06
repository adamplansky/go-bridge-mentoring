package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const password = "secret"

func getPwd() []byte {
	return []byte(password)
}

func genPassword(pwd []byte) ([]byte, error) {
	// bcrypt automatically salt password
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(hash))
	return hash, nil
}

func main() {
	hash, err := genPassword(getPwd())
	fmt.Println(hash, err)
	genPassword(getPwd())

	genPassword(getPwd())

}
