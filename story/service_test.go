package story_test

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
	"github.com/AdhityaRamadhanus/chronicle/story"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	storyService    story.Service
	storyRepository *postgre.StoryRepository
	topicRepository *postgre.TopicRepository
	// specific test case var
	storyId int
	topics  chronicle.Topics
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

	_, err = db.Query(
		`INSERT INTO topics (name, slug, createdat, updatedat)
		VALUES ('Pemilu 2019', 'pemilu-2019', now(), now()),
		('Pilkada 2019', 'pilkada-2019', now(), now()),
		('Jokowi 2019', 'jokowi-2019', now(), now()),
		('Prabowo 2019', 'prabowo-2019', now(), now())`)

	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in filling topics table"))
	}
}

func TestMain(m *testing.M) {
	// log.SetLevel(log.WarnLevel)
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
	storyRepository = postgre.NewStoryRepository(db, "stories")
	topicRepository = postgre.NewTopicRepository(db, "topics")

	storyService = story.NewService(storyRepository)

	code := m.Run()
	os.Exit(code)
}

func TestCreateStoryIntegration(t *testing.T) {
	topics, _, err := topicRepository.All(chronicle.PagingOptions{
		Limit:  3,
		Offset: 0,
		SortBy: "createdAt",
		Order:  "asc",
	})

	if err != nil {
		t.Error("Failed to get topics to for stories", err)
	}

	stories := chronicle.Stories{
		chronicle.Story{
			Title: "Dikalahkan Jepang, Timnas U-19 Gagal ke Piala Dunia",
			Slug:  "dikalahkan-jepang-timnas-u-19-gagal-ke-piala-dunia",
			Content: `Asa Timnas U-19 Indonesia mentas di Piala Dunia U-20 2019 pupus. 
						 Berlaga melawan Jepang pada perempat final Piala Asia U-19 2018 di Stadion Ut`,
			Reporter: "Adhitya Ramadhanus",
			Editor:   "Adhitya Ramadhanus",
			Author:   "Adhitya Ramadhanus",
			Media:    []byte("{}"),
			Excerpt:  "Timnas Gagal melaju ke pialla dunia",
			Status:   chronicle.StoryPublishStatus,
			Topics:   topics,
		},
		chronicle.Story{
			Title:    "Test aja",
			Slug:     "test-aja",
			Content:  "Bertiga melawan jepang pada perempat final",
			Reporter: "Adhitya Ramadhanus",
			Editor:   "Adhitya Ramadhanus",
			Author:   "Adhitya Ramadhanus",
			Media:    []byte("{}"),
			Status:   chronicle.StoryDraftStatus,
			Excerpt:  "Timnas Gagal melaju ke pialla dunia",
			Topics:   topics[1:3],
		},
	}

	for _, story := range stories {
		createdStory, err := storyService.CreateStory(story)

		// take one topic, save its id to test getTopicByID later
		storyId = createdStory.ID

		if err != nil {
			t.Error("Failed to create story", err)
		}

		assert.Equal(t, createdStory.Content, story.Content)
	}
}

func TestGetStoriesIntegration(t *testing.T) {
	testCases := []struct {
		ExpectedStoriesCount int
		ExpectedStoriesSlugs []string
		FilterOption         chronicle.StoryFilterOptions
		PagingOption         chronicle.PagingOptions
	}{
		{
			ExpectedStoriesCount: 2,
			ExpectedStoriesSlugs: []string{
				"test-aja",
				"dikalahkan-jepang-timnas-u-19-gagal-ke-piala-dunia",
			},
			FilterOption: chronicle.StoryFilterOptions{},
			PagingOption: chronicle.PagingOptions{
				SortBy: "createdAt",
				Order:  "desc",
				Limit:  2,
				Offset: 0,
			},
		},
		{
			ExpectedStoriesCount: 1,
			ExpectedStoriesSlugs: []string{
				"test-aja",
			},
			FilterOption: chronicle.StoryFilterOptions{
				Status: chronicle.StoryDraftStatus,
			},
			PagingOption: chronicle.PagingOptions{
				SortBy: "createdAt",
				Order:  "asc",
				Limit:  1,
				Offset: 0,
			},
		},
	}

	for _, testCase := range testCases {
		stories, _, err := storyService.GetStories(testCase.FilterOption, testCase.PagingOption)
		if err != nil {
			t.Error("Failed to create topic", err)
		}
		assert.Equal(t, len(stories), testCase.ExpectedStoriesCount)

		for idx, story := range stories {
			assert.Equal(t, story.Slug, testCase.ExpectedStoriesSlugs[idx])
		}
	}
}

func TestGetStoryByIDIntegration(t *testing.T) {
	testCases := []struct {
		StoryId           int
		ExpectedError     bool
		ExpectedErrorType error
	}{
		{
			StoryId:       storyId,
			ExpectedError: false,
		},
		{
			StoryId:           storyId + 1,
			ExpectedError:     true,
			ExpectedErrorType: story.ErrNoStoryFound,
		},
	}

	for _, test := range testCases {
		_, err := storyService.GetStoryByID(test.StoryId)

		if test.ExpectedError {
			assert.EqualError(t, err, test.ExpectedErrorType.Error(), "Should return ErrStoryNotFound")
		} else {
			if err != nil {
				t.Error("Failed to get story by id", err)
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
			Slug:          "dikalahkan-jepang-timnas-u-19-gagal-ke-piala-dunia",
			ExpectedError: false,
		},
		{
			Slug:              "testaja",
			ExpectedError:     true,
			ExpectedErrorType: story.ErrNoStoryFound,
		},
	}

	for _, test := range testCases {
		_, err := storyService.GetStoryBySlug(test.Slug)

		if test.ExpectedError {
			assert.EqualError(t, err, test.ExpectedErrorType.Error(), "Should return ErrStoryNotFound")
		} else {
			if err != nil {
				t.Error("Failed to get story by slug", err)
			}
		}
	}
}

func TestUpdateStoryIntegration(t *testing.T) {
	story, err := storyRepository.Find(storyId)
	if err != nil {
		t.Error("Failed to get story to update", err)
	}

	story.Status = chronicle.StoryPublishStatus
	story.Editor = "Bukan Adhitya Ramadhanus"

	updatedStory, err := storyService.UpdateStory(story)

	if err != nil {
		t.Error("Failed to update story", err)
	}

	assert.Equal(t, updatedStory.Status, story.Status)
	assert.Equal(t, updatedStory.Editor, story.Editor)
}

func TestDeleteStoryIntegration(t *testing.T) {
	if err := storyService.DeleteStoryByID(storyId); err != nil {
		t.Error("Failed to delete story", err)
	}
}
