package postgre

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	chronicle "github.com/adhityaramadhanus/chronicle"
	function "github.com/adhityaramadhanus/chronicle/function"
)

/*
StoryRepository is implementation of StoryRepository interface
of chronicle domain using postgre
*/
type StoryRepository struct {
	db *sqlx.DB
}

//NewStoryRepository is constructor to create story repository
func NewStoryRepository(conn *sqlx.DB, tableName string) *StoryRepository {
	return &StoryRepository{
		db: conn,
	}
}

//Find find story by id
func (s StoryRepository) Find(id int) (story chronicle.Story, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Find))
		}
	}()

	story = chronicle.Story{}
	query := `SELECT
							id,
							title, 
							slug, 
							excerpt, 
							content,
							reporter,
							editor,
							author,
							status,
							media, 
							likes,
							shares,
							views,
							createdAt, 
							updatedAt
						FROM stories 
						WHERE id=$1`

	err = s.db.Get(&story, query, id)
	if err != nil {
		return chronicle.Story{}, err
	}

	// fill Topics

	story.Topics, err = s.getTopicsForStory(story.ID)
	if err != nil {
		return chronicle.Story{}, err
	}

	return story, nil
}

//FindBySlug find story by slug
func (s StoryRepository) FindBySlug(slug string) (story chronicle.Story, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.FindBySlug))
		}
	}()

	story = chronicle.Story{}
	query := `SELECT
							id,
							title, 
							slug, 
							excerpt, 
							content,
							reporter,
							editor,
							author,
							status,
							media, 
							likes,
							shares,
							views,
							createdAt, 
							updatedAt
						FROM stories 
						WHERE slug=$1`

	err = s.db.Get(&story, query, slug)
	if err != nil {
		return chronicle.Story{}, err
	}

	// fill s.getTopicsForStory(story.ID)
	story.Topics, err = s.getTopicsForStory(story.ID)
	if err != nil {
		return chronicle.Story{}, err
	}

	return story, err
}

//Delete delete story by id
func (s StoryRepository) Delete(id int) (err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Delete))
		}
	}()

	query := `DELETE FROM stories where id=$1`

	deleteStatement, err := s.db.Prepare(query)
	if err != nil {
		return err
	}

	defer deleteStatement.Close()
	_, err = deleteStatement.Exec(id)
	return err
}

//FindByStatus find all story with status x
func (s StoryRepository) FindByStatus(status string, option chronicle.PagingOptions) (stories chronicle.Stories, storiesCount int, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.FindByStatus))
		}
	}()

	stories = chronicle.Stories{}
	selectQuery := fmt.Sprintf(
		`SELECT
			id,
			title, 
			slug, 
			excerpt,
			author,
			status,
			media, 
			likes,
			shares,
			views,
			createdAt, 
			updatedAt
		FROM stories
		WHERE status=$1
		ORDER BY %s %s 
		LIMIT %d 
		OFFSET %d`,
		option.SortBy,
		option.Order,
		option.Limit,
		option.Offset,
	)

	err = s.db.Select(&stories, selectQuery, status)
	if err != nil {
		return chronicle.Stories{}, 0, err
	}

	// count topics for pagination
	countQuery := `SELECT count(*) FROM stories WHERE status=$1`
	row := s.db.QueryRow(countQuery, status)
	row.Scan(&storiesCount)

	if len(stories) == 0 {
		return chronicle.Stories{}, storiesCount, nil
	}

	// fill Topics
	if err := s.getTopicsForStories(&stories); err != nil {
		return chronicle.Stories{}, 0, err
	}

	return stories, storiesCount, nil
}

//FindByTopicAndStatus find all story with topic x and status y
func (s StoryRepository) FindByTopicAndStatus(topic int, status string, option chronicle.PagingOptions) (stories chronicle.Stories, storiesCount int, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.FindByTopicAndStatus))
		}
	}()

	// building where statement and query arguments
	queryArgs := []interface{}{topic}
	whereStatement := "WHERE topic_stories.topicId=$1"
	if status != "" {
		whereStatement += "AND stories.status=$2"
		queryArgs = append(queryArgs, status)
	}

	stories = chronicle.Stories{}
	selectQuery := fmt.Sprintf(
		`SELECT
			stories.id,
			stories.title, 
			stories.slug, 
			stories.excerpt,
			stories.author,
			stories.status,
			stories.media, 
			stories.likes,
			stories.shares,
			stories.views,
			stories.createdat,
			stories.updatedat
		FROM topic_stories
		INNER JOIN stories ON (topic_stories.storyId = stories.id) 
		%s
		ORDER BY %s %s 
		LIMIT %d 
		OFFSET %d`,
		whereStatement,
		option.SortBy,
		option.Order,
		option.Limit,
		option.Offset,
	)

	err = s.db.Select(&stories, selectQuery, queryArgs...)
	if err != nil {
		return chronicle.Stories{}, 0, err
	}

	if len(stories) == 0 {
		return chronicle.Stories{}, 0, nil
	}

	countQuery := fmt.Sprintf(
		`SELECT count(distinct storyId) 
		FROM topic_stories 
		INNER JOIN stories ON (topic_stories.storyId = stories.id)  
		%s`,
		whereStatement,
	)

	row := s.db.QueryRow(countQuery, queryArgs...)
	row.Scan(&storiesCount)

	// fill Topics
	if err := s.getTopicsForStories(&stories); err != nil {
		return chronicle.Stories{}, 0, err
	}

	return stories, storiesCount, nil
}

//All get all stories
func (s StoryRepository) All(option chronicle.PagingOptions) (stories chronicle.Stories, storiesCount int, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.All))
		}
	}()

	stories = chronicle.Stories{}
	selectQuery := fmt.Sprintf(
		`SELECT
			id,
			title, 
			slug, 
			excerpt,
			author,
			status,
			media, 
			likes,
			shares,
			views,
			createdAt, 
			updatedAt
		FROM stories 
		ORDER BY %s %s 
		LIMIT %d 
		OFFSET %d`,
		option.SortBy,
		option.Order,
		option.Limit,
		option.Offset,
	)

	err = s.db.Select(&stories, selectQuery)
	if err != nil {
		return chronicle.Stories{}, 0, err
	}

	countQuery := `SELECT count(*) FROM stories`
	row := s.db.QueryRow(countQuery)
	row.Scan(&storiesCount)

	if len(stories) == 0 {
		return chronicle.Stories{}, storiesCount, nil
	}

	// fill Topics
	if err := s.getTopicsForStories(&stories); err != nil {
		return chronicle.Stories{}, 0, err
	}

	return stories, storiesCount, nil
}

//Insert insert story to datastore
func (s StoryRepository) Insert(story chronicle.Story) (createdStory chronicle.Story, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Insert))
		}
	}()

	query := `INSERT INTO stories (
							title, 
							slug, 
							excerpt, 
							content,
							reporter,
							editor,
							author,
							status,
							media, 
							createdAt, 
							updatedAt
						) VALUES (
							:title, 
							:slug, 
							:excerpt, 
							:content,
							:reporter,
							:editor,
							:author,
							:status,
							:media, 
							now(), 
							now()
						) RETURNING id`

	rows, err := s.db.NamedQuery(query, story)
	if err != nil {
		return chronicle.Story{}, err
	}

	if rows.Next() {
		rows.Scan(&story.ID)
	}

	if err := s.setTopicsForStory(story.ID, story.Topics); err != nil {
		return chronicle.Story{}, err
	}

	return s.Find(story.ID)
}

//Update update story
func (s StoryRepository) Update(story chronicle.Story) (createdStory chronicle.Story, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Update))
		}
	}()

	query := `UPDATE stories SET (
							title, 
							slug, 
							excerpt, 
							content,
							reporter,
							editor,
							author,
							status,
							media, 
							updatedAt
						) = (
							:title, 
							:slug, 
							:excerpt, 
							:content,
							:reporter,
							:editor,
							:author,
							:status,
							:media, 
							now()
						) WHERE id=:id`

	_, err = s.db.NamedQuery(query, story)
	if err != nil {
		return chronicle.Story{}, err
	}

	return s.Find(story.ID)
}

// internal function
func (s StoryRepository) setTopicsForStory(storyId int, topics chronicle.Topics) (err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.setTopicsForStory))
		}
	}()

	if len(topics) == 0 {
		return nil
	}

	multipleValueStatement := []string{}
	for _, topic := range topics {
		multipleValueStatement = append(
			multipleValueStatement,
			fmt.Sprintf("(%d, %d, now(), now())", storyId, topic.ID),
		)
	}
	// insert topics to junction table
	topicQuery := fmt.Sprintf(
		`INSERT INTO topic_stories (
			storyId,
			topicId,
			createdAt,
			updatedAt
		)
		VALUES %s`,
		strings.Join(multipleValueStatement, ","),
	)
	// no transactions
	_, err = s.db.Exec(topicQuery)
	return err
}

func (s StoryRepository) getTopicsForStory(storyId int) (topics chronicle.Topics, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.getTopicsForStory))
		}
	}()

	// fill Topics
	topicQuery := `SELECT
									topics.id,
									topics.name, 
									topics.slug, 
									topics.createdat, 
									topics.updatedat
								FROM topic_stories
								INNER JOIN topics ON (topic_stories.topicid = topics.id)
								WHERE topic_stories.storyid=$1`

	topics = chronicle.Topics{}
	err = s.db.Select(&topics, topicQuery, storyId)
	return topics, err
}

func (s StoryRepository) getTopicsForStories(stories *chronicle.Stories) (err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.getTopicsForStories))
		}
	}()

	storyIds := []interface{}{}
	storyIdsPlaceholders := []string{}
	storyTopics := map[int]chronicle.Topics{}
	for idx, story := range *stories {
		storyIds = append(storyIds, story.ID)
		storyIdsPlaceholders = append(storyIdsPlaceholders, fmt.Sprintf("$%d", idx+1))
		storyTopics[story.ID] = chronicle.Topics{}
	}

	topicQuery := fmt.Sprintf(
		`SELECT
			topic_stories.topicId,
			topic_stories.storyId,
			topics.id,
			topics.name,
			topics.slug,
			topics.createdat,
			topics.updatedat
		FROM topic_stories
		INNER JOIN topics ON (topic_stories.topicId = topics.id)
		WHERE topic_stories.storyid in (%s)`, strings.Join(storyIdsPlaceholders, ","))
	rows, err := s.db.Queryx(topicQuery, storyIds...)
	if err != nil {
		return err
	}

	for rows.Next() {
		topic := chronicle.Topic{}
		var storyId int
		var topicId int
		rows.Scan(
			&topicId,
			&storyId,
			&topic.ID,
			&topic.Name,
			&topic.Slug,
			&topic.CreatedAt,
			&topic.UpdatedAt,
		)
		storyTopics[storyId] = append(storyTopics[storyId], topic)
	}

	for idx, story := range *stories {
		(*stories)[idx].Topics = storyTopics[story.ID]
	}

	return nil
}
