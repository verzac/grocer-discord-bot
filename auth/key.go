package auth

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const validChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"
const tokenLen = 32

var maxInt = big.NewInt(int64(len(validChars) - 1))

func GenerateKey() (string, error) {
	o := ""
	for i := 0; i < tokenLen; i++ {
		n, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			return "", err
		}
		o += string(validChars[n.Int64()])
	}
	return o, nil
}

func HashKey(key string) (string, error) {
	hashedT, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedT), nil
}
