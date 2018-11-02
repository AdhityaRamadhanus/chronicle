package story

import (
	"database/sql"
	"strconv"

	"github.com/adhityaramadhanus/chronicle"
	"github.com/adhityaramadhanus/chronicle/function"
	"github.com/pkg/errors"
)

var (
	//ErrNoStoryFound sub-domain specific error
	ErrNoStoryFound = errors.New("Cannot find Story")
)

//Service provide an interface to story domain service
type Service interface {
	CreateStory(story chronicle.Story) (createdStory chronicle.Story, err error)
	UpdateStory(story chronicle.Story) (updatedStory chronicle.Story, err error)
	GetStories(filter chronicle.StoryFilterOptions, option chronicle.PagingOptions) (chronicle.Stories, int, error)
	GetStoryByID(id int) (chronicle.Story, error)
	GetStoryBySlug(slug string) (chronicle.Story, error)
	DeleteStoryByID(id int) error
}

func NewService(storyRepository chronicle.StoryRepository) Service {
	return &service{
		storyRepository: storyRepository,
	}
}

type service struct {
	storyRepository chronicle.StoryRepository
}

func (s *service) CreateStory(story chronicle.Story) (createdStory chronicle.Story, err error) {
	defer func() {
		if err != nil && err != ErrNoStoryFound {
			err = errors.Wrap(err, function.GetFunctionName(s.CreateStory))
		}
	}()

	return s.storyRepository.Insert(story)
}

func (s *service) UpdateStory(story chronicle.Story) (updatedStory chronicle.Story, err error) {
	defer func() {
		if err != nil && err != ErrNoStoryFound {
			err = errors.Wrap(err, function.GetFunctionName(s.UpdateStory))
		}
	}()

	return s.storyRepository.Update(story)
}

func (s *service) GetStories(filter chronicle.StoryFilterOptions, option chronicle.PagingOptions) (stories chronicle.Stories, storiesCount int, err error) {
	defer func() {
		if err != nil && err != ErrNoStoryFound {
			err = errors.Wrap(err, function.GetFunctionName(s.GetStories))
		}
	}()

	// filter by status only
	if filter.Status != "" && filter.Topic == "" {
		return s.storyRepository.FindByStatus(filter.Status, option)
	}

	// filter by topic only
	if filter.Topic != "" {
		topicId, _ := strconv.Atoi(filter.Topic)
		return s.storyRepository.FindByTopicAndStatus(topicId, filter.Status, option)
	}

	return s.storyRepository.All(option)
}

func (s *service) GetStoryByID(id int) (story chronicle.Story, err error) {
	defer func() {
		if err != nil && err != ErrNoStoryFound {
			err = errors.Wrap(err, function.GetFunctionName(s.GetStoryByID))
		}
	}()

	story, err = s.storyRepository.Find(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return story, ErrNoStoryFound
		default:
			return story, err
		}
	}

	return story, nil
}

func (s *service) GetStoryBySlug(slug string) (story chronicle.Story, err error) {
	defer func() {
		if err != nil && err != ErrNoStoryFound {
			err = errors.Wrap(err, function.GetFunctionName(s.GetStoryBySlug))
		}
	}()

	story, err = s.storyRepository.FindBySlug(slug)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return story, ErrNoStoryFound
		default:
			return story, err
		}
	}

	return story, nil
}

func (s *service) DeleteStoryByID(id int) (err error) {
	defer func() {
		if err != nil && err != ErrNoStoryFound {
			err = errors.Wrap(err, function.GetFunctionName(s.DeleteStoryByID))
		}
	}()

	return s.storyRepository.Delete(id)
}
