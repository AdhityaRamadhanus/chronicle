package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdhityaRamadhanus/chronicle/config"
	"github.com/AdhityaRamadhanus/chronicle/server"
	"github.com/AdhityaRamadhanus/chronicle/server/handlers"
	"github.com/AdhityaRamadhanus/chronicle/storage/postgre"
	_redis "github.com/AdhityaRamadhanus/chronicle/storage/redis"
	"github.com/AdhityaRamadhanus/chronicle/story"
	"github.com/AdhityaRamadhanus/chronicle/topic"
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
	pgConnString := postgre.GetConnString()

	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"database":      "postgres",
		"database-name": viper.GetString("database.dbname"),
		"host":          viper.GetString("database.host"),
		"port":          viper.GetString("database.port"),
	}).Info("Connected to database")

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")),
		Password: viper.GetString("redis.password"), // no password set
		DB:       viper.GetInt("redis.db"),          // use default DB
	})

	log.WithFields(log.Fields{
		"cache-server": "redis",
		"host":         viper.GetString("redis.host"),
		"port":         viper.GetString("redis.port"),
	}).Info("Connected to cache-server")

	_, err = redisClient.Ping().Result()
	if err != nil {
		os.Setenv("cache_response", "false")
		log.WithError(err).Error("Failed to connect to redis, caching response is disabled")
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

	log.WithField("Port", server.Port).Info("Chronicle API Server is running")
	log.Fatal(srv.ListenAndServe())
}
