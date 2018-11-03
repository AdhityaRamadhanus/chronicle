package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/AdhityaRamadhanus/chronicle/config"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	godotenv.Load()
	config.Init(os.Getenv("ENV"), []string{})
}

func main() {
	pgConnString := fmt.Sprintf(`
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

	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"database": "postgres",
		"host":     viper.GetString("database.host"),
		"port":     viper.GetString("database.port"),
	}).Info("Connected to postgres")

	migrationDir := "storage/postgre/migration"
	migrationFiles := []string{}
	files, _ := ioutil.ReadDir(migrationDir)
	for _, file := range files {
		if !file.IsDir() {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Sort(sort.StringSlice(migrationFiles))
	for _, migrationFile := range migrationFiles {
		log.Info("Running ", migrationFile)
		filePath := path.Join(migrationDir, migrationFile)
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		fileBytes, _ := ioutil.ReadAll(file)
		file.Close()

		sqlQuery := string(fileBytes)
		_, err = db.Queryx(sqlQuery)
		if err != nil {
			log.Fatal(err)
		}
	}
}
