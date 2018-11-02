package chronicle

import "time"

//Topic is domain entity
type Topic struct {
	ID        int
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

//Topics short way to define array of story
type Topics []Topic

//TopicRepository provide an interface to get topic entities
type TopicRepository interface {
	Find(id int) (Topic, error)
	FindBySlug(slug string) (Topic, error)
	All(option PagingOptions) (topics Topics, topicsCount int, err error)
	Insert(topic Topic) (createdTopic Topic, err error)
	Update(topic Topic) (updatedTopic Topic, err error)
	Delete(id int) error
}
