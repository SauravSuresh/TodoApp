package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var secretKey = []byte("secretpassword")

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s %v \n", msg, err)
	}
}

func GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{}                            // Creates a claims object
	claims["user_id"] = userID                           // adds to the claims map a user_id string to userID uint
	claims["exp"] = time.Now().Add(time.Hour + 1).Unix() // adds to the claims map a exp token of 1 hr

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // creates a new token with the claims created above
	return token.SignedString(secretKey)                       // signs the token with the secret key
}

func VerifyToken(tokenstring string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("invalid signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
