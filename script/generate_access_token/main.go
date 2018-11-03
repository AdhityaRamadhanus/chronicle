package script

import (
	"log"
	"os"
	"time"

	"github.com/adhityaramadhanus/chronicle/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	godotenv.Load()
	config.Init(os.Getenv("ENV"), []string{})
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client":    "Kumkum",
		"timestamp": time.Now(),
	})
	tokenString, err := jwtToken.SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tokenString)
}
