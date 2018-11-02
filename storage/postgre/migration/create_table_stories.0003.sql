CREATE TABLE IF NOT EXISTS stories (
  id serial PRIMARY KEY,
  title varchar(255) NOT NULL,
  slug varchar(255) NOT NULL,
  excerpt varchar(255) NOT NULL,
	content text NOT NULL,
	reporter  varchar(25),
	editor  varchar(25),
	author  varchar(25),
	media json,
	status      VARCHAR(20) NOT NULL,
	likes     int DEFAULT 0,
	shares     int DEFAULT 0,
	views     int DEFAULT 0,
  createdAt TIMESTAMP,
  updatedAt TIMESTAMP,

	CONSTRAINT stories_unique_slug UNIQUE (slug)
);

CREATE INDEX index_stories_on_status ON public.stories USING btree (status) ;
CREATE INDEX index_stories_on_slug ON public.stories USING btree (slug) ;
CREATE INDEX index_stories_on_updatedAt ON public.stories USING btree (updatedAt) ;
