package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"

	"github.com/AdhityaRamadhanus/chronicle"
	"github.com/AdhityaRamadhanus/chronicle/server/internal/contextkey"
	"github.com/AdhityaRamadhanus/chronicle/server/middlewares"
	"github.com/AdhityaRamadhanus/chronicle/server/render"
	"github.com/AdhityaRamadhanus/chronicle/story"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type StoryHandler struct {
	StoryService story.Service
	CacheService chronicle.CacheService
}

func (h StoryHandler) RegisterRoutes(router *mux.Router) {
	authMiddleware := middlewares.Authenticate
	cacheMiddleware := middlewares.Cache(h.CacheService)

	router.HandleFunc("/stories/", authMiddleware(cacheMiddleware("60s", h.getStories))).Methods("GET")
	router.HandleFunc("/stories/insert", authMiddleware(h.createStory)).Methods("POST")

	router.HandleFunc("/stories/{id:[0-9]+}", authMiddleware(cacheMiddleware("60s", h.getStoryByID))).Methods("GET")
	router.HandleFunc("/stories/{id:[0-9]+}/update", authMiddleware(h.updateStory)).Methods("PATCH")
	router.HandleFunc("/stories/{id:[0-9]+}/delete", authMiddleware(h.deleteStoryByID)).Methods("DELETE")

	router.HandleFunc("/stories/{slug}", authMiddleware(cacheMiddleware("60s", h.getStoryBySlug))).Methods("GET")
}

func (h *StoryHandler) getStories(res http.ResponseWriter, req *http.Request) {
	// Pagination
	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}
	page, _ := strconv.Atoi(req.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	sortby := req.URL.Query().Get("sort-by")
	if sortby == "" {
		sortby = "updatedAt"
	}
	order := req.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	// filter
	status := req.URL.Query().Get("status")
	topic := req.URL.Query().Get("topic")

	getStoriesRequest := struct {
		Limit  int    `valid:"int"`
		Page   int    `valid:"int"`
		Order  string `valid:"in(asc|desc)"`
		SortBy string `valid:"in(createdAt|updatedAt)"`
		Status string `valid:"in(Draft|Deleted|Publish)"`
		Topic  string `valid:"int"`
	}{
		Limit:  limit,
		Page:   page,
		SortBy: sortby,
		Order:  order,
		Status: status,
		Topic:  topic,
	}

	if ok, err := govalidator.ValidateStruct(getStoriesRequest); !ok || err != nil {
		RenderError(res, ErrInvalidRequest, err.Error())
		return
	}

	stories, storiesCount, err := h.StoryService.GetStories(
		chronicle.StoryFilterOptions{
			Status: status,
			Topic:  topic,
		},
		chronicle.PagingOptions{
			Limit:  limit,
			Offset: (page - 1) * limit,
			SortBy: sortby,
			Order:  order,
		},
	)

	if err != nil {
		log.WithFields(log.Fields{
			"request":      getStoriesRequest,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Getting Stories")

		RenderError(res, ErrSomethingWrong)
		return
	}

	totalPage := int(math.Ceil(float64(storiesCount) / float64(limit)))
	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status":  http.StatusOK,
		"stories": stories,
		"pagination": map[string]interface{}{
			"totalItems":   storiesCount,
			"page":         page,
			"itemsPerPage": limit,
			"totalPage":    totalPage,
		},
	})
}

func (h *StoryHandler) createStory(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	createStoryRequest := struct {
		TopicIDs []int           `json:"topics" valid:"-"`
		Media    json.RawMessage `json:"media" valid:"-"`
		Title    string          `json:"title" valid:"required"`
		Excerpt  string          `json:"excerpt" valid:"required"`
		Content  string          `json:"content" valid:"required"`
		Reporter string          `json:"reporter" valid:"required"`
		Editor   string          `json:"editor" valid:"required"`
		Author   string          `json:"author" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &createStoryRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(createStoryRequest); !ok || err != nil {
		RenderError(res, ErrInvalidRequest, err.Error())
		return
	}

	newStoryTopics := chronicle.Topics{}
	for _, topicId := range createStoryRequest.TopicIDs {
		newStoryTopics = append(newStoryTopics, chronicle.Topic{ID: topicId})
	}

	newStory := chronicle.Story{
		Topics:   newStoryTopics,
		Media:    createStoryRequest.Media,
		Title:    createStoryRequest.Title,
		Slug:     chronicle.Slugify(createStoryRequest.Title),
		Content:  createStoryRequest.Content,
		Excerpt:  createStoryRequest.Excerpt,
		Reporter: createStoryRequest.Reporter,
		Editor:   createStoryRequest.Editor,
		Author:   createStoryRequest.Author,
		Status:   chronicle.StoryDraftStatus,
	}

	createdStory, err := h.StoryService.CreateStory(newStory)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      createStoryRequest,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Creating Stories")

		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"status": http.StatusCreated,
		"story":  createdStory,
	})
}

func (h *StoryHandler) updateStory(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	updateStoryRequest := struct {
		Status   string          `json:"status" valid:"in(Draft|Deleted|Publish)"`
		Media    json.RawMessage `json:"media" valid:"-"`
		Title    string          `json:"title"`
		Excerpt  string          `json:"excerpt"`
		Content  string          `json:"content"`
		Reporter string          `json:"reporter"`
		Editor   string          `json:"editor"`
		Author   string          `json:"author"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &updateStoryRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(updateStoryRequest); !ok || err != nil {
		RenderError(res, ErrInvalidRequest, err.Error())
		return
	}

	storyId, _ := strconv.Atoi(params["id"])
	foundStory, err := h.StoryService.GetStoryByID(storyId)

	if err != nil && err == story.ErrNoStoryFound {
		render.JSON(res, http.StatusNotFound, map[string]interface{}{
			"status": http.StatusNotFound,
			"error": map[string]interface{}{
				"code":    "ErrNoStoryFound",
				"message": err.Error(),
			},
		})
		return
	}

	if updateStoryRequest.Media != nil {
		foundStory.Media = updateStoryRequest.Media
	}

	if updateStoryRequest.Title != "" {
		foundStory.Title = updateStoryRequest.Title
		foundStory.Slug = chronicle.Slugify(updateStoryRequest.Title)
	}

	if updateStoryRequest.Excerpt != "" {
		foundStory.Excerpt = updateStoryRequest.Excerpt
	}

	if updateStoryRequest.Content != "" {
		foundStory.Content = updateStoryRequest.Content
	}

	if updateStoryRequest.Reporter != "" {
		foundStory.Reporter = updateStoryRequest.Reporter
	}

	if updateStoryRequest.Editor != "" {
		foundStory.Editor = updateStoryRequest.Editor
	}

	if updateStoryRequest.Author != "" {
		foundStory.Author = updateStoryRequest.Author
	}

	if updateStoryRequest.Status != "" {
		foundStory.Status = updateStoryRequest.Status
	}

	updatedStory, err := h.StoryService.UpdateStory(foundStory)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      updateStoryRequest,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Updating Story")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"story":  updatedStory,
	})
}

func (h *StoryHandler) getStoryByID(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	storyId, _ := strconv.Atoi(params["id"])
	foundStory, err := h.StoryService.GetStoryByID(storyId)

	if err != nil && err == story.ErrNoStoryFound {
		render.JSON(res, http.StatusNotFound, map[string]interface{}{
			"status": http.StatusNotFound,
			"error": map[string]interface{}{
				"code":    "ErrNoStoryFound",
				"message": err.Error(),
			},
		})
		return
	}

	if err != nil {
		log.WithFields(log.Fields{
			"request":      storyId,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Story By ID")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"story":  foundStory,
	})
}

func (h *StoryHandler) deleteStoryByID(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	storyId, _ := strconv.Atoi(params["id"])
	err := h.StoryService.DeleteStoryByID(storyId)

	if err != nil {
		log.WithFields(log.Fields{
			"request":      storyId,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Delete Story by ID")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status":  http.StatusOK,
		"message": "Story Deleted",
	})
}

func (h *StoryHandler) getStoryBySlug(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	slug := params["slug"]
	foundStory, err := h.StoryService.GetStoryBySlug(slug)

	if err != nil && err == story.ErrNoStoryFound {
		render.JSON(res, http.StatusNotFound, map[string]interface{}{
			"status": http.StatusNotFound,
			"error": map[string]interface{}{
				"code":    "ErrNoStoryFound",
				"message": err.Error(),
			},
		})
		return
	}

	if err != nil {
		log.WithFields(log.Fields{
			"request":      slug,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Get Story By Slug")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"story":  foundStory,
	})
}
