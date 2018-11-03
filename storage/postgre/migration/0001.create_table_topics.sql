CREATE TABLE IF NOT EXISTS topics (
  id serial PRIMARY KEY,
  name varchar(255) NOT NULL,
  slug varchar(255) NOT NULL,
  createdAt TIMESTAMP,
  updatedAt TIMESTAMP,

	CONSTRAINT topics_unique_slug UNIQUE (slug)
);

CREATE INDEX index_topics_on_slug ON public.topics USING btree (slug);