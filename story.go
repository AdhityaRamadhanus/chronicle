package chronicle

import (
	"encoding/json"
	"time"
)

var (
	//StoryDraftStatus provide a uniform way to use draft status instead of literal string
	StoryDraftStatus = "Draft"
	//StoryDeletedStatus provide a uniform way to use deleted status instead of literal string
	StoryDeletedStatus = "Deleted"
	//StoryPublishStatus provide a uniform way to use publish status instead of literal string
	StoryPublishStatus = "Publish"
)

//Story is domain entity
type Story struct {
	ID       int
	Media    json.RawMessage
	Title    string
	Slug     string
	Excerpt  string
	Content  string
	Reporter string
	Editor   string
	Author   string
	Status   string
	Topics   Topics

	// stats
	Likes     int
	Shares    int
	Views     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

//Stories short way to define array of story
type Stories []Story

//StoryFilterOptions struct used as parameter to filter story by status and its topic
type StoryFilterOptions struct {
	Status string
	Topic  string
}

//StoryRepository provide an interface to get story entities
type StoryRepository interface {
	Find(id int) (Story, error)
	FindBySlug(slug string) (Story, error)
	FindByStatus(status string, option PagingOptions) (stories Stories, storiesCount int, err error)
	FindByTopicAndStatus(topic int, status string, option PagingOptions) (stories Stories, storiesCount int, err error)
	All(option PagingOptions) (stories Stories, storiesCount int, err error)
	Insert(story Story) (createdStory Story, err error)
	Update(story Story) (updatedStory Story, err error)
	Delete(id int) error
}
