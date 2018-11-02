package topic

import (
	"database/sql"

	"github.com/pkg/errors"
	"gitlab.com/adhityaramadhanus/chronicle"
)

var (
	//ErrNoTopicFound sub-domain specific error
	ErrNoTopicFound = errors.New("Cannot find Topic")
)

//Service provide an interface to topic domain service
type Service interface {
	CreateTopic(topic chronicle.Topic) (createdTopic chronicle.Topic, err error)
	UpdateTopic(topic chronicle.Topic) (updatedTopic chronicle.Topic, err error)
	GetTopics(option chronicle.TopicPageOptions) (chronicle.Topics, int, error)
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
	return s.topicRepository.Insert(topic)
}

func (s *service) UpdateTopic(topic chronicle.Topic) (updatedTopic chronicle.Topic, err error) {
	return s.topicRepository.Update(topic)
}

func (s *service) GetTopics(option chronicle.TopicPageOptions) (chronicle.Topics, int, error) {
	return s.topicRepository.All(option)
}

func (s *service) GetTopicByID(id int) (chronicle.Topic, error) {
	topic, err := s.topicRepository.Find(id)
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

func (s *service) GetTopicBySlug(slug string) (chronicle.Topic, error) {
	story, err := s.topicRepository.FindBySlug(slug)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return story, ErrNoTopicFound
		default:
			return story, err
		}
	}

	return story, nil
}

func (s *service) DeleteTopicByID(id int) error {
	return s.topicRepository.Delete(id)
}
