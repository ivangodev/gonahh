package psql

import (
	"database/sql"
	"fmt"
	"github.com/ivangodev/gonahh/entity"
	_ "github.com/lib/pq"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type PSql struct {
	Db *sql.DB
}

func OpenDB() (*sql.DB, error) {
	host := os.Getenv("DATABASE_HOSTNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := 5432
	user := "postgres"

	time.Sleep(5 * time.Second)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func NewPSql(db *sql.DB) *PSql {
	q := `
    CREATE TABLE url (
        job_id integer PRIMARY KEY,
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
    );`
	_, err := db.Exec(q)
	if err != nil {
		if !strings.HasSuffix(err.Error(), "already exists") {
			panic(fmt.Sprintf("Failed to create tables: %s", err))
		}
	}

	return &PSql{Db: db}
}

func (r *PSql) deleteTables() error {
	q := `DROP TABLE url, name, engwords, category CASCADE;`
	_, err := r.Db.Exec(q)
	return err
}

func (r *PSql) GetJobsNumber(jobName string) (jobsNumber int, err error) {
	query := `SELECT COUNT(*) as jobsNumber FROM name WHERE name LIKE $1`
	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	rows, err := r.Db.Query(query, likeJobName)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		err = rows.Scan(&jobsNumber)
		if err != nil {
			return 0, err
		}
	}
	return jobsNumber, nil
}

func (r *PSql) getTop(query string, params ...interface{}) ([]entity.KeywordRate, error) {
	rows, err := r.Db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	res := make([]entity.KeywordRate, 0)
	for rows.Next() {
		var kr entity.KeywordRate
		err = rows.Scan(&kr.Keyword, &kr.Rate)
		if err != nil {
			return nil, err
		}
		res = append(res, kr)
	}

	return res, nil
}

func (r *PSql) GetCategories(jobName string) ([]string, error) {
	q := `
    SELECT DISTINCT category
    FROM category JOIN engwords USING(word) JOIN name USING(job_id)
    WHERE name LIKE $1
    `
	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	rows, err := r.Db.Query(q, likeJobName)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0)
	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, nil
}

func (r *PSql) GetCategoryTop(jobName, category string) (entity.CategoryTop, error) {
	jobsNumber, err := r.GetJobsNumber(jobName)
	if err != nil {
		return entity.CategoryTop{}, err
	}
	if jobsNumber == 0 {
		return entity.CategoryTop{
			Category: category,
			Top:      make([]entity.KeywordRate, 0),
		}, nil
	}

	q := `
    SELECT word, ROUND((COUNT(word)*100.0 / $1), 0) as rate
    FROM engwords JOIN name USING(job_id) JOIN category USING(word)
    WHERE name LIKE $2 AND category = $3
    GROUP BY word ORDER BY rate DESC
    `
	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	top, err := r.getTop(q, jobsNumber, likeJobName, category)
	if err != nil {
		return entity.CategoryTop{}, nil
	}
	return entity.CategoryTop{Category: category, Top: top}, nil
}

func (r *PSql) GetAllTop(jobName string) ([]entity.KeywordRate, error) {
	jobsNumber, err := r.GetJobsNumber(jobName)
	if err != nil {
		return nil, err
	}
	res := make([]entity.KeywordRate, 0)
	if jobsNumber == 0 {
		return res, nil
	}

	q := `
    SELECT word, ROUND((COUNT(word)*100.0 / $1), 0) as rate
    FROM engwords JOIN name USING(job_id)
    WHERE name LIKE $2
    GROUP BY word ORDER BY rate DESC
    `
	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	return r.getTop(q, jobsNumber, likeJobName)
}

func (r *PSql) Write(data entity.Schema, w io.Writer) error {
	urlId := make(map[string]int)
	var res string
	var id int

	for url, info := range data.URLtoJobInfo {
		id++
		urlId[url] = id
		res += fmt.Sprintf("INSERT INTO url(job_id, url) VALUES (%d,'%s');\n",
			id, url)
		res += fmt.Sprintf("INSERT INTO name(job_id, name) VALUES (%d,'%s');\n",
			id, info.Name)
		for _, keyword := range info.Keywords {
			res += fmt.Sprintf("INSERT INTO engwords VALUES (%d, '%s');\n",
				id, keyword)
		}
	}

	for keyword, category := range data.KeywordCategory {
		res += fmt.Sprintf("INSERT INTO category VALUES ('%s', '%s');\n",
			keyword, category)
	}

	_, err := io.WriteString(w, res)
	return err
}

func (repo *PSql) Read(r io.Reader) error {
	q, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("Failed to read from reader: %w", err)
	}

	_, err = repo.Db.Exec(string(q))
	if err != nil {
		return fmt.Errorf("Failed to exec query: %s: %w", string(q), err)
	}
	return nil
}
