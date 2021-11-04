\c test

DROP TABLE url cascade ;
DROP TABLE name ;
DROP TABLE engwords ;

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
	word VARCHAR ( 255 ) NOT NULL
);
