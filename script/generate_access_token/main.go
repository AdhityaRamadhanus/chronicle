package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/AdhityaRamadhanus/chronicle/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	godotenv.Load()
	config.Init(os.Getenv("ENV"), []string{})

	clientName := flag.String("client", "chronicle-app", "client name")
	flag.Parse()
	if (*clientName) == "" {
		log.Println("Please input client name")
	}

	log.Println("Generating access token for ", *clientName)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client":    *clientName,
		"timestamp": time.Now(),
	})
	tokenString, err := jwtToken.SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tokenString)
}
