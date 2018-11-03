package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adhityaramadhanus/chronicle/config"
	"github.com/adhityaramadhanus/chronicle/server"
	"github.com/adhityaramadhanus/chronicle/server/handlers"
	"github.com/adhityaramadhanus/chronicle/storage/postgre"
	_redis "github.com/adhityaramadhanus/chronicle/storage/redis"
	"github.com/adhityaramadhanus/chronicle/story"
	"github.com/adhityaramadhanus/chronicle/topic"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sebest/logrusly"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	logruslyHook *logrusly.LogglyHook
)

func init() {
	godotenv.Load()
	config.Init(os.Getenv("ENV"), []string{})

	switch os.Getenv("ENV") {
	case "production":
		logruslyHook = logrusly.NewLogglyHook(
			viper.GetString("logglytoken"),
			viper.GetString("logglyhost"),
			log.WarnLevel,
			"chronicle",
		)

		// set log
		log.SetFormatter(&log.JSONFormatter{})
		log.SetLevel(log.WarnLevel)
		log.AddHook(logruslyHook)
	default:
		log.SetOutput(os.Stdout)
	}
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

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")),
		Password: viper.GetString("redis.password"), // no password set
		DB:       viper.GetInt("redis.db"),          // use default DB
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	// Repositories
	storyRepository := postgre.NewStoryRepository(db, "stories")
	topicRepository := postgre.NewTopicRepository(db, "topics")

	storyService := story.NewService(storyRepository)
	topicService := topic.NewService(topicRepository)
	cacheService := _redis.NewCacheService(redisClient)

	storyHandler := handlers.StoryHandler{
		StoryService: storyService,
		CacheService: cacheService,
	}
	topicHandler := handlers.TopicHandler{
		TopicService: topicService,
		CacheService: cacheService,
	}
	handlers := []server.Handler{
		storyHandler,
		topicHandler,
	}
	server := server.NewServer(handlers)
	srv := server.CreateHttpServer()

	// Handle SIGINT, SIGTERN, SIGHUP signal from OS
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-termChan
		log.Warn("Receiving signal, Shutting down server")
		srv.Close()
	}()

	log.WithField("URL", server.Addr).Info("Chronicle API Server is running")
	log.Fatal(srv.ListenAndServe())
}
