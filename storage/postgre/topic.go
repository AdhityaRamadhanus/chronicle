package postgre

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	chronicle "github.com/adhityaramadhanus/chronicle"
	function "github.com/adhityaramadhanus/chronicle/function"
)

/*
TopicRepository is implementation of TopicRepository interface
of chronicle domain using postgre
*/
type TopicRepository struct {
	db *sqlx.DB
}

//NewTopicRepository is constructor to create topic repository
func NewTopicRepository(conn *sqlx.DB, tableName string) *TopicRepository {
	return &TopicRepository{
		db: conn,
	}
}

//Find find topic by id
func (s TopicRepository) Find(id int) (topic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Find))
		}
	}()

	topic = chronicle.Topic{}
	query := `SELECT
							id,
							name,
							slug,
							createdAt,
							updatedAt
						FROM topics 
						WHERE id=$1`

	err = s.db.Get(&topic, query, id)
	return topic, err
}

//FindBySlug find topic by slug
func (s TopicRepository) FindBySlug(slug string) (topic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Find))
		}
	}()

	topic = chronicle.Topic{}
	query := `SELECT
							id,
							name,
							slug,
							createdAt,
							updatedAt
						FROM topics 
						WHERE slug=$1`

	err = s.db.Get(&topic, query, slug)
	return topic, err
}

//Delete delete topic by id
func (s TopicRepository) Delete(id int) (err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Delete))
		}
	}()

	query := `DELETE FROM topics where id=$1`

	deleteStatement, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer deleteStatement.Close()
	_, err = deleteStatement.Exec(id)
	return err
}

//All get all topic
func (s TopicRepository) All(option chronicle.PagingOptions) (topics chronicle.Topics, topicsCount int, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.All))
		}
	}()

	topics = chronicle.Topics{}
	selectQuery := fmt.Sprintf(
		`SELECT
			id,
			name,
			slug,
			createdAt,
			updatedAt
		FROM topics 
		ORDER BY %s %s 
		LIMIT %d 
		OFFSET %d`,
		option.SortBy,
		option.Order,
		option.Limit,
		option.Offset,
	)

	err = s.db.Select(&topics, selectQuery)
	if err != nil {
		return chronicle.Topics{}, 0, err
	}

	countQuery := `SELECT count(*) FROM topics`
	row := s.db.QueryRow(countQuery)
	err = row.Scan(&topicsCount)

	return topics, topicsCount, err
}

//Insert insert topic to datastore
func (s TopicRepository) Insert(topic chronicle.Topic) (createdTopic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Insert))
		}
	}()

	query := `INSERT INTO topics (
							name,
							slug,
							createdAt, 
							updatedAt
						) VALUES (
							:name, 
							:slug, 
							now(), 
							now()
						) RETURNING id`

	rows, err := s.db.NamedQuery(query, topic)
	if err != nil {
		return chronicle.Topic{}, err
	}

	if rows.Next() {
		rows.Scan(&topic.ID)
	}

	return s.Find(topic.ID)
}

//Update update topic
func (s TopicRepository) Update(topic chronicle.Topic) (updatedTopic chronicle.Topic, err error) {
	defer func() {
		if err != nil && err != sql.ErrNoRows {
			err = errors.Wrap(err, function.GetFunctionName(s.Update))
		}
	}()

	query := `UPDATE topics SET (
							name,
							slug,
							updatedAt
						) = (
							:name, 
							:slug, 
							now()
						) WHERE id=:id`

	_, err = s.db.NamedQuery(query, topic)
	if err != nil {
		return chronicle.Topic{}, err
	}

	return s.Find(topic.ID)
}
