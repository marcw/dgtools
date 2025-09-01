-- +goose Up
CREATE TABLE IF NOT EXISTS discogs_artists (
    id integer NOT NULL,
    name character varying,
    real_name character varying,
    profile text,
    data_quality character varying,
    name_variations jsonb DEFAULT '[]'::jsonb,
    urls jsonb DEFAULT '[]'::jsonb
);
CREATE TABLE IF NOT EXISTS discogs_artists_aliases (
    artist_id integer NOT NULL,
    alias_id integer NOT NULL
);
CREATE TABLE IF NOT EXISTS discogs_artists_members (
    artist_id integer NOT NULL,
    member_id integer NOT NULL
);
CREATE TABLE IF NOT EXISTS discogs_labels (
    id integer NOT NULL,
    name character varying,
    profile text,
    contact_info text,
    data_quality character varying,
    parent_label_id integer,
    urls jsonb
);
CREATE TABLE IF NOT EXISTS discogs_master_artists (
    master_id integer NOT NULL,
    artist_id integer NOT NULL,
    name character varying,
    name_variation character varying,
    "join" character varying
);
CREATE TABLE IF NOT EXISTS discogs_masters (
    id integer NOT NULL,
    main_release_id integer,
    title character varying NOT NULL,
    year integer,
    data_quality character varying,
    videos jsonb,
    genres jsonb,
    styles jsonb,
    series jsonb
);
CREATE TABLE IF NOT EXISTS discogs_release_artists (
    release_id integer NOT NULL,
    artist_id integer NOT NULL,
    name character varying,
    name_variation character varying,
    "join" character varying
);
CREATE TABLE IF NOT EXISTS discogs_release_extra_artists (
    release_id integer NOT NULL,
    artist_id integer NOT NULL,
    name character varying,
    name_variation character varying,
    role character varying
);
CREATE TABLE IF NOT EXISTS discogs_release_labels (
    release_id integer NOT NULL,
    label_id integer NOT NULL,
    name character varying,
    catno character varying
);
CREATE TABLE IF NOT EXISTS discogs_releases (
    id integer NOT NULL,
    master_id integer,
    is_main_release boolean,
    title character varying,
    status character varying,
    country character varying,
    released character varying,
    data_quality character varying,
    genres jsonb,
    styles jsonb,
    videos jsonb,
    identifiers jsonb,
    tracklist jsonb,
    formats jsonb,
    companies jsonb,
    series jsonb,
    year integer,
    thumb character varying,
    cover_image character varying,
    notes text
);
ALTER TABLE ONLY discogs_artists
ADD CONSTRAINT discogs_artists_pkey PRIMARY KEY (id);
ALTER TABLE ONLY discogs_labels
ADD CONSTRAINT discogs_labels_pkey PRIMARY KEY (id);
ALTER TABLE ONLY discogs_masters
ADD CONSTRAINT discogs_masters_pkey PRIMARY KEY (id);
ALTER TABLE ONLY discogs_releases
ADD CONSTRAINT discogs_releases_pkey PRIMARY KEY (id);
CREATE INDEX IF NOT EXISTS index_discogs_artists_aliases_on_alias_id_and_artist_id ON discogs_artists_aliases USING btree (alias_id, artist_id);
CREATE INDEX IF NOT EXISTS index_discogs_artists_aliases_on_artist_id_and_alias_id ON discogs_artists_aliases USING btree (artist_id, alias_id);
CREATE INDEX IF NOT EXISTS index_discogs_artists_members_on_artist_id_and_member_id ON discogs_artists_members USING btree (artist_id, member_id);
CREATE INDEX IF NOT EXISTS index_discogs_artists_members_on_member_id_and_artist_id ON discogs_artists_members USING btree (member_id, artist_id);
CREATE INDEX IF NOT EXISTS index_discogs_artists_on_id ON discogs_artists USING btree (id);
CREATE INDEX IF NOT EXISTS index_discogs_labels_on_id ON discogs_labels USING btree (id);
CREATE INDEX IF NOT EXISTS index_discogs_labels_on_parent_label_id ON discogs_labels USING btree (parent_label_id);
CREATE INDEX IF NOT EXISTS index_discogs_master_artists_on_artist_id_and_master_id ON discogs_master_artists USING btree (artist_id, master_id);
CREATE INDEX IF NOT EXISTS index_discogs_master_artists_on_master_id_and_artist_id ON discogs_master_artists USING btree (master_id, artist_id);
CREATE UNIQUE INDEX index_discogs_masters_on_id ON discogs_masters USING btree (id);
CREATE INDEX IF NOT EXISTS index_discogs_masters_on_main_release_id ON discogs_masters USING btree (main_release_id);
CREATE INDEX IF NOT EXISTS index_discogs_release_artists_on_artist_id_and_release_id ON discogs_release_artists USING btree (artist_id, release_id);
CREATE INDEX IF NOT EXISTS index_discogs_release_artists_on_release_id_and_artist_id ON discogs_release_artists USING btree (release_id, artist_id);
CREATE INDEX IF NOT EXISTS index_discogs_release_extra_artists_on_artist_id_and_release_id ON discogs_release_extra_artists USING btree (artist_id, release_id);
CREATE INDEX IF NOT EXISTS index_discogs_release_extra_artists_on_release_id_and_artist_id ON discogs_release_extra_artists USING btree (release_id, artist_id);
CREATE INDEX IF NOT EXISTS index_discogs_release_labels_on_label_id_and_release_id ON discogs_release_labels USING btree (label_id, release_id);
CREATE INDEX IF NOT EXISTS index_discogs_release_labels_on_release_id_and_label_id ON discogs_release_labels USING btree (release_id, label_id);
CREATE INDEX IF NOT EXISTS index_discogs_releases_on_id ON discogs_releases USING btree (id);
CREATE INDEX IF NOT EXISTS index_discogs_releases_on_master_id ON discogs_releases USING btree (master_id);
CREATE INDEX IF NOT EXISTS index_discogs_releases_on_title_and_year ON discogs_releases USING btree (title, year);
-- +goose Down
DROP TABLE discogs_artists;
DROP TABLE discogs_artists_aliases;
DROP TABLE discogs_artists_members;
DROP TABLE discogs_labels;
DROP TABLE discogs_master_artists;
DROP TABLE discogs_masters;
DROP TABLE discogs_release_artists;
DROP TABLE discogs_release_extra_artists;
DROP TABLE discogs_release_labels;
DROP TABLE discogs_releases;
