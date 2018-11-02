CREATE TABLE IF NOT EXISTS topic_stories (
  topicId int REFERENCES topics(id) ON DELETE CASCADE,
  storyId int REFERENCES stories(id) ON DELETE CASCADE,
  createdAt TIMESTAMP,
  updatedAt TIMESTAMP,

  CONSTRAINT topic_stories_pkey PRIMARY KEY (topicId, storyId)
);

CREATE INDEX index_topic_stories_on_topicId ON public.topic_stories USING btree (topicId);