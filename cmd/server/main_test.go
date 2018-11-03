package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/chronicle"
	"github.com/AdhityaRamadhanus/chronicle/config"
	cs "github.com/AdhityaRamadhanus/chronicle/server"
	"github.com/AdhityaRamadhanus/chronicle/server/handlers"
	"github.com/AdhityaRamadhanus/chronicle/storage/postgre"
	_redis "github.com/AdhityaRamadhanus/chronicle/storage/redis"
	"github.com/AdhityaRamadhanus/chronicle/story"
	"github.com/AdhityaRamadhanus/chronicle/topic"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type (
	DefautInnerError struct {
		Code    string
		Message string
	}
	DefaultErrorBody struct {
		Status int
		Error  DefautInnerError
	}
	DefaultInnerPagination struct {
		TotalItems   int
		Page         int
		ItemsPerPage int
		TotalPage    int
	}

	DetailTopicBody struct {
		Status int
		Topic  chronicle.Topic
	}

	ListTopicsBody struct {
		Status     int
		Topics     chronicle.Topics
		Pagination DefaultInnerPagination
	}

	DetailStoryBody struct {
		Status int
		Story  chronicle.Story
	}

	ListStoriesBody struct {
		Status     int
		Stories    chronicle.Stories
		Pagination DefaultInnerPagination
	}
)

var (
	server      *http.Server
	accessToken string
	// cross test variable
	topicId int
	storyId int
)

func decodeResponseJSON(t *testing.T, response *httptest.ResponseRecorder, decodedResponse interface{}) error {
	jsonContentTypeHeader := "application/json; charset=utf-8"
	requestContentTypeHeader := response.Header().Get("Content-Type")
	if jsonContentTypeHeader != requestContentTypeHeader {
		return errors.New("Expected Content Type JSON")
	}

	return json.NewDecoder(response.Body).Decode(&decodedResponse)
}

func createHttpJSONRequest(method string, path string, requestBody interface{}) (*http.Request, error) {
	var httpReq *http.Request
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		jsonReqBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to marshal post request body")
		}
		httpReq, err = http.NewRequest(method, path, bytes.NewBuffer(jsonReqBody))
		if err != nil {
			return nil, errors.Wrap(err, "Failed to create post request body")
		}
		httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	case "GET":
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to marshal get request body")
		}
		httpReq = req
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	return httpReq, nil
}

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
	if err := config.Init("testing", []string{"../../config/testing"}); err != nil {
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

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")),
		Password: viper.GetString("redis.password"), // no password set
		DB:       viper.GetInt("redis.db"),          // use default DB
	})

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

	handlers := []cs.Handler{
		storyHandler,
		topicHandler,
	}
	server = cs.NewServer(handlers).CreateHttpServer()

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client":    "chronicle-test",
		"timestamp": time.Now(),
	})
	tokenString, err := jwtToken.SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		log.Fatal(err)
	}

	accessToken = tokenString
	os.Setenv("cache_response", "false")

	code := m.Run()
	os.Exit(code)
}

// TEST TOPICS ENDPOINT

func TestCreateTopicsIntegration(t *testing.T) {
	url := "/api/topics/insert"
	method := "POST"

	t.Logf("Testing %s %s", method, url)
	testCases := []struct {
		RequestBody          interface{}
		ExpectedStatus       int
		ExpectedResponseBody interface{}
	}{
		{
			RequestBody: map[string]interface{}{
				"name": "Pemilu 2019",
			},
			ExpectedStatus: 201,
		},
		{
			RequestBody: map[string]interface{}{
				"name": "Pilkada 2019",
			},
			ExpectedStatus: 201,
		},
		{
			RequestBody:    map[string]interface{}{},
			ExpectedStatus: 422,
		},
	}

	for _, test := range testCases {
		request, err := createHttpJSONRequest(method, url, test.RequestBody)
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailTopicBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
			topicId = requestBody.Topic.ID
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestGetTopicsIntegration(t *testing.T) {
	baseUrl := "/api/topics/"
	method := "GET"

	testCases := []struct {
		ExpectedStatus       int
		Querystring          string
		ExpectedTopicsCount  int
		ExpectedTopicsLength int
	}{
		{
			Querystring:          "page=1&order=asc&sort-by=updatedAt&limit=20",
			ExpectedStatus:       200,
			ExpectedTopicsCount:  2,
			ExpectedTopicsLength: 2,
		},
		{
			Querystring:          "page=2&order=asc&sort-by=updatedAt&limit=20",
			ExpectedStatus:       200,
			ExpectedTopicsCount:  2,
			ExpectedTopicsLength: 0,
		},
		{
			Querystring:    "page=1&order=xxxx&sort-by=updatedAt&limit=20",
			ExpectedStatus: 422,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("?%s", test.Querystring)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := ListTopicsBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
			assert.Equal(t, requestBody.Pagination.TotalItems, test.ExpectedTopicsCount, fmt.Sprintf("Should return %d topics count", test.ExpectedTopicsCount))
			assert.Equal(t, len(requestBody.Topics), test.ExpectedTopicsLength, fmt.Sprintf("Should return %d topics", test.ExpectedTopicsLength))
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestGetTopicBySlugIntegration(t *testing.T) {
	baseUrl := "/api/topics"
	method := "GET"

	testCases := []struct {
		ExpectedStatus int
		Slug           string
	}{
		{
			Slug:           "pemilu-2019",
			ExpectedStatus: 200,
		},
		{
			Slug:           "test",
			ExpectedStatus: 404,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%s", test.Slug)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailTopicBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestGetTopicByIDIntegration(t *testing.T) {
	baseUrl := "/api/topics"
	method := "GET"

	testCases := []struct {
		ExpectedStatus int
		ID             int
	}{
		{
			ID:             topicId,
			ExpectedStatus: 200,
		},
		{
			ID:             topicId + 1,
			ExpectedStatus: 404,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%d", test.ID)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailTopicBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestUpdateTopicIntegration(t *testing.T) {
	baseUrl := "/api/topics"
	method := "PATCH"

	testCases := []struct {
		ExpectedStatus int
		ID             int
		RequestBody    interface{}
	}{
		{
			RequestBody: map[string]interface{}{
				"name": "Pemilih 2019",
			},
			ID:             topicId,
			ExpectedStatus: 200,
		},
		{
			RequestBody: map[string]interface{}{
				"name": "Pemilih 2019",
			},
			ID:             topicId + 1,
			ExpectedStatus: 404,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%d/update", test.ID)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, test.RequestBody)
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailTopicBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestDeleteTopicIntegration(t *testing.T) {
	baseUrl := "/api/topics"
	method := "DELETE"

	testCases := []struct {
		ExpectedStatus int
		ID             int
	}{
		{
			ID:             topicId - 1,
			ExpectedStatus: 200,
		},
		{
			ID:             topicId + 1,
			ExpectedStatus: 200,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%d/delete", test.ID)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailTopicBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

// STORIES ENDPOINT TESTS

func TestCreateStoriesIntegration(t *testing.T) {
	url := "/api/stories/insert"
	method := "POST"

	t.Logf("Testing %s %s", method, url)
	testCases := []struct {
		RequestBody          map[string]interface{}
		ExpectedStatus       int
		ExpectedResponseBody interface{}
	}{
		{
			RequestBody: map[string]interface{}{
				"title": "Dikalahkan Jepang, Timnas U-19 Gagal ke Piala Dunia",
				"content": `Asa Timnas U-19 Indonesia mentas di Piala Dunia U-20 2019 pupus. 
							 Berlaga melawan Jepang pada perempat final Piala Asia U-19 2018 di Stadion Ut`,
				"reporter": "Adhitya Ramadhanus",
				"editor":   "Adhitya Ramadhanus",
				"author":   "Adhitya Ramadhanus",
				"media":    "{}",
				"excerpt":  "Timnas Gagal melaju ke pialla dunia",
				"topics":   []int{topicId},
			},
			ExpectedStatus: 201,
		},
		{
			RequestBody: map[string]interface{}{
				"title": "Deutschland Uber Alles",
				"content": `Asa Timnas U-19 Indonesia mentas di Piala Dunia U-20 2019 pupus. 
							 Berlaga melawan Jepang pada perempat final Piala Asia U-19 2018 di Stadion Ut`,
				"reporter": "Adhitya Ramadhanus",
				"editor":   "Adhitya Ramadhanus",
				"author":   "Adhitya Ramadhanus",
				"media":    "{}",
				"excerpt":  "Timnas Gagal melaju ke pialla dunia",
			},
			ExpectedStatus: 201,
		},
		{
			RequestBody: map[string]interface{}{
				"content": `Asa Timnas U-19 Indonesia mentas di Piala Dunia U-20 2019 pupus. 
							 Berlaga melawan Jepang pada perempat final Piala Asia U-19 2018 di Stadion Ut`,
				"reporter": "Adhitya Ramadhanus",
				"editor":   "Adhitya Ramadhanus",
				"author":   "Adhitya Ramadhanus",
				"media":    "{}",
			},
			ExpectedStatus: 422,
		},
	}

	for _, test := range testCases {
		request, err := createHttpJSONRequest(method, url, test.RequestBody)
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailStoryBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
			storyId = requestBody.Story.ID
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestGetStoriesIntegration(t *testing.T) {
	baseUrl := "/api/stories/"
	method := "GET"

	testCases := []struct {
		ExpectedStatus        int
		Querystring           string
		ExpectedStoriesCount  int
		ExpectedStoriesLength int
	}{
		{
			Querystring:           "status=Draft&page=1&order=asc&sort-by=updatedAt&limit=20",
			ExpectedStatus:        200,
			ExpectedStoriesCount:  2,
			ExpectedStoriesLength: 2,
		},
		{
			Querystring:           "page=2&order=asc&sort-by=updatedAt&limit=20",
			ExpectedStatus:        200,
			ExpectedStoriesCount:  2,
			ExpectedStoriesLength: 0,
		},
		{
			Querystring:    "page=1&order=xxxx&sort-by=updatedAt&limit=20",
			ExpectedStatus: 422,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("?%s", test.Querystring)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := ListStoriesBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
			assert.Equal(t, requestBody.Pagination.TotalItems, test.ExpectedStoriesCount, fmt.Sprintf("Should return %d stories count", test.ExpectedStoriesCount))
			assert.Equal(t, len(requestBody.Stories), test.ExpectedStoriesLength, fmt.Sprintf("Should return %d stories", test.ExpectedStoriesLength))
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestGetStoryBySlugIntegration(t *testing.T) {
	baseUrl := "/api/stories"
	method := "GET"

	testCases := []struct {
		ExpectedStatus int
		Slug           string
	}{
		{
			Slug:           "deutschland-uber-alles",
			ExpectedStatus: 200,
		},
		{
			Slug:           "test",
			ExpectedStatus: 404,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%s", test.Slug)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailStoryBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestGetStoryByIDIntegration(t *testing.T) {
	baseUrl := "/api/stories"
	method := "GET"

	testCases := []struct {
		ExpectedStatus int
		ID             int
	}{
		{
			ID:             storyId,
			ExpectedStatus: 200,
		},
		{
			ID:             storyId + 1,
			ExpectedStatus: 404,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%d", test.ID)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailStoryBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestUpdateStoryIntegration(t *testing.T) {
	baseUrl := "/api/stories"
	method := "PATCH"

	testCases := []struct {
		ExpectedStatus int
		ID             int
		RequestBody    interface{}
	}{
		{
			RequestBody: map[string]interface{}{
				"status": "Publish",
			},
			ID:             storyId,
			ExpectedStatus: 200,
		},
		{
			RequestBody: map[string]interface{}{
				"name": "Pemilih 2019",
			},
			ID:             storyId + 1,
			ExpectedStatus: 404,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%d/update", test.ID)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, test.RequestBody)
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailStoryBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}

func TestDeleteStoryIntegration(t *testing.T) {
	baseUrl := "/api/stories"
	method := "DELETE"

	testCases := []struct {
		ExpectedStatus int
		ID             int
	}{
		{
			ID:             storyId - 1,
			ExpectedStatus: 200,
		},
		{
			ID:             storyId + 1,
			ExpectedStatus: 200,
		},
	}

	for _, test := range testCases {
		url := baseUrl + fmt.Sprintf("/%d/delete", test.ID)
		t.Logf("Testing %s %s", method, url)
		request, err := createHttpJSONRequest(method, url, map[string]interface{}{})
		assert.NoError(t, err, "Expected No Error in create request")

		response := httptest.NewRecorder()
		server.Handler.ServeHTTP(response, request)
		assert.Equal(t, test.ExpectedStatus, response.Code, fmt.Sprintf("Expected to return %d", test.ExpectedStatus))

		if response.Code == 200 || response.Code == 201 {
			requestBody := DetailStoryBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		} else {
			requestBody := DefaultErrorBody{}
			err = decodeResponseJSON(t, response, &requestBody)
			assert.NoError(t, err, "Expected No Error in decode response")
		}
	}
}
