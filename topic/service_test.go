package topic_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/AdhityaRamadhanus/chronicle"
	"github.com/AdhityaRamadhanus/chronicle/config"
	"github.com/AdhityaRamadhanus/chronicle/storage/postgre"
	"github.com/AdhityaRamadhanus/chronicle/topic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	topicService    topic.Service
	topicRepository *postgre.TopicRepository

	// specific test case var
	topicId int
)

// make test kind of idempotent
func setupDatabase(db *sqlx.DB) {
	_, err := db.Query("DELETE FROM topic_stories")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from topic_stories"))
	}

	_, err = db.Query("DELETE FROM stories")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from stories"))
	}

	_, err = db.Query("DELETE FROM topics")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from topics"))
	}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.WarnLevel)
	if err := config.Init("testing", []string{"../config/testing"}); err != nil {
		log.Fatal(err)
	}

	pgConnString := fmt.Sprintf(
		`host=%s 
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

	setupDatabase(db)

	// Repositories
	topicRepository = postgre.NewTopicRepository(db, "topics")
	topicService = topic.NewService(topicRepository)

	code := m.Run()
	os.Exit(code)
}

func TestCreateTopicIntegration(t *testing.T) {
	topics := chronicle.Topics{
		chronicle.Topic{
			Name: "Pemilu 2019",
			Slug: "pemilu-2019",
		},
		chronicle.Topic{
			Name: "Pilkada 2019",
			Slug: "pilkada-2019",
		},
		chronicle.Topic{
			Name: "Jokowi 2019",
			Slug: "jokowi-2019",
		},
		chronicle.Topic{
			Name: "Prabowo 2019",
			Slug: "prabowo-2019",
		},
		chronicle.Topic{
			Name: "Sepakbola Dalam Negeri",
			Slug: "sepakbola-dalam-negeri",
		},
	}

	for _, topic := range topics {
		createdTopic, err := topicService.CreateTopic(topic)

		// take one topic, save its id to test getTopicByID later
		topicId = createdTopic.ID

		if err != nil {
			t.Error("Failed to create topic", err)
		}

		assert.Equal(t, createdTopic.Name, topic.Name)
	}
}

func TestGetTopicsIntegration(t *testing.T) {
	testCases := []struct {
		ExpectedTopicsCount int
		ExpectedTopicsSlugs []string
		PagingOption        chronicle.PagingOptions
	}{
		{
			ExpectedTopicsCount: 2,
			ExpectedTopicsSlugs: []string{
				"pemilu-2019",
				"pilkada-2019",
			},
			PagingOption: chronicle.PagingOptions{
				SortBy: "createdAt",
				Order:  "asc",
				Limit:  2,
				Offset: 0,
			},
		},
		{
			ExpectedTopicsCount: 2,
			ExpectedTopicsSlugs: []string{
				"sepakbola-dalam-negeri",
				"prabowo-2019",
			},
			PagingOption: chronicle.PagingOptions{
				SortBy: "createdAt",
				Order:  "desc",
				Limit:  2,
				Offset: 0,
			},
		},
		{
			ExpectedTopicsCount: 2,
			ExpectedTopicsSlugs: []string{
				"pilkada-2019",
				"pemilu-2019",
			},
			PagingOption: chronicle.PagingOptions{
				SortBy: "createdAt",
				Order:  "desc",
				Limit:  3,
				Offset: 3,
			},
		},
	}

	for _, testCase := range testCases {
		topics, _, err := topicService.GetTopics(testCase.PagingOption)
		if err != nil {
			t.Error("Failed to create topic", err)
		}
		assert.Equal(t, len(topics), testCase.ExpectedTopicsCount)

		for idx, topic := range topics {
			assert.Equal(t, topic.Slug, testCase.ExpectedTopicsSlugs[idx])
		}
	}
}

func TestGetTopicByIDIntegration(t *testing.T) {
	testCases := []struct {
		TopicId           int
		ExpectedError     bool
		ExpectedErrorType error
	}{
		{
			TopicId:       topicId,
			ExpectedError: false,
		},
		{
			TopicId:           topicId + 1,
			ExpectedError:     true,
			ExpectedErrorType: topic.ErrNoTopicFound,
		},
	}

	for _, test := range testCases {
		_, err := topicService.GetTopicByID(test.TopicId)

		if test.ExpectedError {
			assert.EqualError(t, err, test.ExpectedErrorType.Error(), "Should return ErrTopicNotFound")
		} else {
			if err != nil {
				t.Error("Failed to get topic by id", err)
			}
		}
	}
}

func TestGetTopicBySlugIntegration(t *testing.T) {
	testCases := []struct {
		Slug              string
		ExpectedError     bool
		ExpectedErrorType error
	}{
		{
			Slug:          "pemilu-2019",
			ExpectedError: false,
		},
		{
			Slug:              "pemilu2019",
			ExpectedError:     true,
			ExpectedErrorType: topic.ErrNoTopicFound,
		},
	}

	for _, test := range testCases {
		_, err := topicService.GetTopicBySlug(test.Slug)

		if test.ExpectedError {
			assert.EqualError(t, err, test.ExpectedErrorType.Error(), "Should return ErrTopicNotFound")
		} else {
			if err != nil {
				t.Error("Failed to get topic by slug", err)
			}
		}
	}
}

func TestUpdateTopicIntegration(t *testing.T) {
	newTopic := chronicle.Topic{
		ID:   topicId,
		Name: "Pemilih 2019",
		Slug: "pemilih-2019",
	}

	updatedTopic, err := topicService.UpdateTopic(newTopic)

	if err != nil {
		t.Error("Failed to get topic by slug", err)
	}

	assert.Equal(t, updatedTopic.Name, newTopic.Name)
	assert.Equal(t, updatedTopic.Slug, newTopic.Slug)
}

func TestDeleteTopicIntegration(t *testing.T) {
	if err := topicService.DeleteTopicByID(topicId); err != nil {
		t.Error("Failed to delete topic", err)
	}
}
