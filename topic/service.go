package topic

import (
	"database/sql"

	"github.com/adhityaramadhanus/chronicle"
	"github.com/adhityaramadhanus/chronicle/function"
	"github.com/pkg/errors"
)

var (
	//ErrNoTopicFound sub-domain specific error
	ErrNoTopicFound = errors.New("Cannot find Topic")
)

//Service provide an interface to topic domain service
type Service interface {
	CreateTopic(topic chronicle.Topic) (createdTopic chronicle.Topic, err error)
	UpdateTopic(topic chronicle.Topic) (updatedTopic chronicle.Topic, err error)
	GetTopics(option chronicle.PagingOptions) (chronicle.Topics, int, error)
	GetTopicByID(id int) (chronicle.Topic, error)
	GetTopicBySlug(slug string) (chronicle.Topic, error)
	DeleteTopicByID(id int) error
}

func NewService(topicRepository chronicle.TopicRepository) Service {
	return &service{
		topicRepository: topicRepository,
	}
}

type service struct {
	topicRepository chronicle.TopicRepository
}

func (s *service) CreateTopic(topic chronicle.Topic) (createdTopic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != ErrNoTopicFound {
			err = errors.Wrap(err, function.GetFunctionName(s.CreateTopic))
		}
	}()

	return s.topicRepository.Insert(topic)
}

func (s *service) UpdateTopic(topic chronicle.Topic) (updatedTopic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != ErrNoTopicFound {
			err = errors.Wrap(err, function.GetFunctionName(s.UpdateTopic))
		}
	}()

	return s.topicRepository.Update(topic)
}

func (s *service) GetTopics(option chronicle.PagingOptions) (topics chronicle.Topics, topicsCount int, err error) {
	defer func() {
		if err != nil && err != ErrNoTopicFound {
			err = errors.Wrap(err, function.GetFunctionName(s.GetTopics))
		}
	}()

	return s.topicRepository.All(option)
}

func (s *service) GetTopicByID(id int) (topic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != ErrNoTopicFound {
			err = errors.Wrap(err, function.GetFunctionName(s.GetTopicByID))
		}
	}()

	topic, err = s.topicRepository.Find(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return topic, ErrNoTopicFound
		default:
			return topic, err
		}
	}

	return topic, nil
}

func (s *service) GetTopicBySlug(slug string) (topic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != ErrNoTopicFound {
			err = errors.Wrap(err, function.GetFunctionName(s.GetTopicBySlug))
		}
	}()

	topic, err = s.topicRepository.FindBySlug(slug)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return topic, ErrNoTopicFound
		default:
			return topic, err
		}
	}

	return topic, nil
}

func (s *service) DeleteTopicByID(id int) (err error) {
	defer func() {
		if err != nil && err != ErrNoTopicFound {
			err = errors.Wrap(err, function.GetFunctionName(s.DeleteTopicByID))
		}
	}()

	return s.topicRepository.Delete(id)
}
