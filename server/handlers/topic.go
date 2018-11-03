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
	topic "github.com/AdhityaRamadhanus/chronicle/topic"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type TopicHandler struct {
	TopicService topic.Service
	CacheService chronicle.CacheService
}

func (h TopicHandler) RegisterRoutes(router *mux.Router) {
	authMiddleware := middlewares.Authenticate
	cacheMiddleware := middlewares.Cache(h.CacheService)

	// bug in gorilla mux, subrouter methods
	router.HandleFunc("/topics/", authMiddleware(cacheMiddleware("60s", h.getTopics))).Methods("GET")
	router.HandleFunc("/topics/insert", authMiddleware(h.createTopic)).Methods("POST")

	router.HandleFunc("/topics/{id:[0-9]+}", authMiddleware(cacheMiddleware("60s", h.getTopicByID))).Methods("GET")
	router.HandleFunc("/topics/{id:[0-9]+}/update", authMiddleware(h.updateTopic)).Methods("PATCH")
	router.HandleFunc("/topics/{id:[0-9]+}/delete", authMiddleware(h.deleteTopicByID)).Methods("DELETE")

	router.HandleFunc("/topics/{slug}", authMiddleware(cacheMiddleware("60s", h.getTopicBySlug))).Methods("GET")
}

func (h *TopicHandler) getTopics(res http.ResponseWriter, req *http.Request) {
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
	order := req.URL.Query().Get("order")

	getTopicsRequest := struct {
		Limit  int    `valid:"int"`
		Page   int    `valid:"int"`
		Order  string `valid:"in(asc|desc), required"`
		SortBy string `valid:"in(createdAt|updatedAt), required"`
	}{
		Limit:  limit,
		Page:   page,
		SortBy: sortby,
		Order:  order,
	}

	if ok, err := govalidator.ValidateStruct(getTopicsRequest); !ok || err != nil {
		RenderError(res, ErrInvalidRequest, err.Error())
		return
	}

	topics, topicsCount, err := h.TopicService.GetTopics(chronicle.PagingOptions{
		Limit:  limit,
		Offset: (page - 1) * limit,
		SortBy: sortby,
		Order:  order,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"request":      getTopicsRequest,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Getting Topics")
		RenderError(res, ErrSomethingWrong)
		return
	}

	totalPage := int(math.Ceil(float64(topicsCount) / float64(limit)))
	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"topics": topics,
		"pagination": map[string]interface{}{
			"totalItems":   topicsCount,
			"page":         page,
			"itemsPerPage": limit,
			"totalPage":    totalPage,
		},
	})
}

func (h *TopicHandler) createTopic(res http.ResponseWriter, req *http.Request) {
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	createTopicRequest := struct {
		Name string `json:"name" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &createTopicRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(createTopicRequest); !ok || err != nil {
		RenderError(res, ErrInvalidRequest, err.Error())
		return
	}

	newTopic := chronicle.Topic{
		Name: createTopicRequest.Name,
		Slug: chronicle.Slugify(createTopicRequest.Name),
	}

	createdTopic, err := h.TopicService.CreateTopic(newTopic)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      createTopicRequest,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Creating Topic")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusCreated, map[string]interface{}{
		"status": http.StatusCreated,
		"topic":  createdTopic,
	})
}

func (h *TopicHandler) getTopicByID(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	topicId, _ := strconv.Atoi(params["id"])
	foundTopic, err := h.TopicService.GetTopicByID(topicId)

	if err != nil && err == topic.ErrNoTopicFound {
		render.JSON(res, http.StatusNotFound, map[string]interface{}{
			"status": http.StatusNotFound,
			"error": map[string]interface{}{
				"code":    "ErrNoTopicFound",
				"message": err.Error(),
			},
		})
		return
	}

	if err != nil {
		log.WithFields(log.Fields{
			"request":      topicId,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Getting Topic By ID")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"topic":  foundTopic,
	})
}

func (h *TopicHandler) updateTopic(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	// Read Body, limit to 1 MB //
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		RenderError(res, ErrFailedToReadBody)
		return
	}

	updateTopicRequest := struct {
		Name string `json:"name" valid:"required"`
	}{}

	// Deserialize
	if err := json.Unmarshal(body, &updateTopicRequest); err != nil {
		RenderError(res, ErrFailedToUnmarshalJSON)
		return
	}

	if err := req.Body.Close(); err != nil {
		RenderError(res, ErrSomethingWrong)
		return
	}

	if ok, err := govalidator.ValidateStruct(updateTopicRequest); !ok || err != nil {
		RenderError(res, ErrInvalidRequest, err.Error())
		return
	}

	topicId, _ := strconv.Atoi(params["id"])
	oldTopic, err := h.TopicService.GetTopicByID(topicId)

	if err != nil && err == topic.ErrNoTopicFound {
		render.JSON(res, http.StatusNotFound, map[string]interface{}{
			"status": http.StatusNotFound,
			"error": map[string]interface{}{
				"code":    "ErrNoTopicFound",
				"message": err.Error(),
			},
		})
		return
	}

	oldTopic.Name = updateTopicRequest.Name
	oldTopic.Slug = chronicle.Slugify(updateTopicRequest.Name)

	updatedTopic, err := h.TopicService.UpdateTopic(oldTopic)
	if err != nil {
		log.WithFields(log.Fields{
			"request":      updateTopicRequest,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Updating Topic")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"data":   updatedTopic,
	})
}

func (h *TopicHandler) deleteTopicByID(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	topicId, _ := strconv.Atoi(params["id"])
	err := h.TopicService.DeleteTopicByID(topicId)

	if err != nil {
		log.WithFields(log.Fields{
			"request":      topicId,
			"client":       req.Context().Value(contextkey.ClientID).(string),
			"x-request-id": req.Header.Get("X-Request-ID"),
		}).WithError(err).Error("Error Handler Delete Topic By ID")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status":  http.StatusOK,
		"message": "Topic Deleted",
	})
}

func (h *TopicHandler) getTopicBySlug(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	slug, _ := params["slug"]
	foundTopic, err := h.TopicService.GetTopicBySlug(slug)

	if err != nil && err == topic.ErrNoTopicFound {
		render.JSON(res, http.StatusNotFound, map[string]interface{}{
			"status": http.StatusNotFound,
			"error": map[string]interface{}{
				"code":    "ErrNoTopicFound",
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
		}).WithError(err).Error("Error Handler Getting Topic By Slug")
		RenderError(res, ErrSomethingWrong)
		return
	}

	render.JSON(res, http.StatusOK, map[string]interface{}{
		"status": http.StatusOK,
		"topic":  foundTopic,
	})
}
