package main

import (
	"flag"
	"fmt"
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
	expiration := flag.String("exp", "24h", "client name")
	flag.Parse()

	nowInSeconds := time.Now().Unix()

	claims := jwt.MapClaims{
		"iss":    "chronicle-api",
		"aud":    *clientName,
		"sub":    fmt.Sprintf("chronicle-access-token|%s|%d", *clientName, nowInSeconds),
		"iat":    nowInSeconds,
		"client": *clientName,
	}

	var exp time.Duration
	exp, err := time.ParseDuration(*expiration)
	if err == nil {
		claims["exp"] = nowInSeconds + int64(exp.Seconds())
	}

	log.Println("Generating access token for ", *clientName)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tokenString)
}
