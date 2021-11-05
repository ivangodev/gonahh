package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vanyaio/gohh/fetcher"
	"github.com/vanyaio/gohh/webapp"
	"os"
	"strings"
)

//TODO: this is awful duplicate of Vacancy. Refactor it to own package.
type Job struct {
	URL      string
	name     string
	engWords []string
}

type category struct {
	name  string
	words []string
}

const (
	lookingURL     = iota
	lookingName    = iota
	lookingEngword = iota
)

const (
	lookingCategoryName = iota
	pickingWord         = iota
)

func showHelp(error_code int) {
	fmt.Println("usage: main  [--web|--read --file path|--fetch --file path]")
	fmt.Println("--web: start web application")
	fmt.Println("--fetch: Fetch vacancies info from API and dump to --file")
	fmt.Println("--read: Read vacancies info from --file and dump to database")
	os.Exit(error_code)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func handleFetch(filename string) {
	vacanciesURLs := fetcher.GetVacanciesURLs()
	vacancies := fetcher.NewVacancies(vacanciesURLs)

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	for _, vacancy := range vacancies {
		vacancy.LogVacancy(f)
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func openDB() *sql.DB {
	host := os.Getenv("DATABASE_HOSTNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	port := 5432
	user := "postgres"
	dbname := "test"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	return db
}

func dumpToDB(job *Job) {
	if job == nil {
		return
	}
	db := openDB()
	defer db.Close()

	stat := `INSERT INTO url(url) VALUES ($1)`
	_, err := db.Exec(stat, job.URL)
	if err != nil {
		return /* It's likely violates duplicates - drop such job. */
	}

	rows, err := db.Query("SELECT job_id FROM url WHERE url='" + job.URL + "'")
	checkErr(err)
	var job_id int
	for rows.Next() {
		err = rows.Scan(&job_id)
		checkErr(err)
	}

	stat = `INSERT INTO name(job_id, name) VALUES ($1, $2)`
	job.name = strings.ToLower(job.name)
	_, err = db.Exec(stat, job_id, job.name)
	checkErr(err)

	for _, w := range job.engWords {
		stat = `INSERT INTO engwords(job_id, word) VALUES ($1, $2)`
		_, err = db.Exec(stat, job_id, w)
		checkErr(err)
	}
}


func replaceSynonym(db *sql.DB, target, src string) {
	q := "SELECT distinct job_id FROM engwords"
	rows, err := db.Query(q)
	checkErr(err)

	for rows.Next() {
		var job_id int
		err = rows.Scan(&job_id)
		checkErr(err)

		q = `UPDATE engwords SET word = $1 WHERE word = $2 and job_id = $3`
		db.Exec(q, target, src, job_id)
		q = `DELETE FROM engwords WHERE word = $1 and job_id = $2`
		db.Exec(q, src, job_id)
	}
}

func handleCleanDB(filename string) {
	file, err := os.Open(filename)
	checkErr(err)
	defer file.Close()

	db := openDB()
	defer db.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()

		q := `DELETE FROM engwords WHERE word=$1`
		_, err = db.Exec(q, word)
		checkErr(err)
	}
	replaceSynonym(db, "javascript", "js")
	replaceSynonym(db, "rails", "ror")
	replaceSynonym(db, "vue", "vue.js")
	replaceSynonym(db, "go", "golang")
	replaceSynonym(db, "postgresql", "postgres")
}

func handleRead(filename string) {
	file, err := os.Open(filename)
	checkErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	state := lookingURL
	var v *Job
	for scanner.Scan() {
		str := scanner.Text()
		switch {
		case str == "":
			dumpToDB(v)
			state = lookingURL
			v = nil
		case state == lookingURL:
			v = new(Job)
			v.engWords = make([]string, 0)
			v.URL = str
			state = lookingName
		case state == lookingName:
			v.name = str
			state = lookingEngword
		case state == lookingEngword:
			v.engWords = append(v.engWords, str)
		}
	}
	dumpToDB(v)

	checkErr(scanner.Err())
}

func setCategory(c *category) {
	db := openDB()
	defer db.Close()

	for _, w := range c.words {
		q := `INSERT INTO category(word, category) VALUES ($1, $2)`
		db.Exec(q, w, c.name)
	}
}

func handleCategorize(filename string) {
	file, err := os.Open(filename)
	checkErr(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	state := lookingCategoryName
	var c *category
	for scanner.Scan() {
		str := scanner.Text()
		switch {
		case str == "":
			setCategory(c)
			state = lookingCategoryName
			c = nil
		case state == lookingCategoryName:
			c = new(category)
			c.name = str
			c.words = make([]string, 0)
			state = pickingWord
		case state == pickingWord:
			c.words = append(c.words, str)
		}
	}
	setCategory(c)
}

func handleWeb() {
	webapp.Main()
}

func main() {
	web := flag.Bool("web", false, "start web app")
	fetch := flag.Bool("fetch", false, "do fetch")
	read := flag.Bool("read", false, "do read")
	cleandb := flag.Bool("cleandb", false, "Clean database from banned words")
	categorize := flag.Bool("categorize", false, "Categorize words")
	help := flag.Bool("help", false, "show help")
	file := flag.String("file", "", "file to fetch to or read from")
	flag.Parse()

	if *help {
		showHelp(0)
	}

	switch {
	case *web:
		handleWeb()
	case *fetch:
		handleFetch(*file)
	case *read:
		handleRead(*file)
	case *cleandb:
		handleCleanDB(*file)
	case *categorize:
		handleCategorize(*file)
	default:
		fmt.Fprintf(os.Stderr, "Wrong input!\n")
		showHelp(1)
	}
}
