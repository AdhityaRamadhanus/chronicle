package postgre

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func GetConnString() string {
	if os.Getenv("ENV") == "production" {
		//heroku specific
		return os.Getenv("DATABASE_URL")
	}

	return fmt.Sprintf(`
		host=%s 
		port=%d 
		user=%s 
		password=%s 
		dbname=%s 
		sslmode=%s`,
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.dbname"),
		viper.GetString("database.sslmode"),
	)
}
