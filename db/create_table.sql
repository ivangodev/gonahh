CREATE DATABASE test;
\c test

CREATE TABLE url (
	job_id SERIAL PRIMARY KEY,
	url VARCHAR ( 255 ) UNIQUE NOT NULL
);

CREATE TABLE name (
	job_id integer PRIMARY KEY REFERENCES url (job_id),
	name VARCHAR ( 255 ) NOT NULL
);


CREATE TABLE engwords (
	job_id INTEGER REFERENCES url (job_id),
	word VARCHAR ( 255 ) NOT NULL,
	UNIQUE (job_id, word)
);

CREATE TABLE category (
	word VARCHAR ( 255 ) NOT NULL UNIQUE,
	category VARCHAR ( 255 ) NOT NULL
);
