package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vanyaio/gohh/fetcher"
	"os"
	"strings"
)

//TODO: this is awful duplicate of Vacancy. Refactor it to own package.
type Job struct {
	URL      string
	name     string
	engWords []string
}

const (
	lookingURL     = iota
	lookingName    = iota
	lookingEngword = iota
)

func showHelp(error_code int) {
	fmt.Println("usage: main --file /path/to/file [--read|--fetch]")
	fmt.Println("Fetch to or read from file vacancies info.")
	fmt.Println("When read, info is dumped to database.")
	os.Exit(error_code)
}

func handleWrongInput(fetch, read bool, file string) {
	var inputIsWrong bool

	if fetch == read {
		fmt.Fprintf(os.Stderr, "use either read or fetch\n")
		inputIsWrong = true
	}

	if file == "" {
		fmt.Fprintf(os.Stderr, "file is mandatory argument\n")
		inputIsWrong = true
	}

	if inputIsWrong {
		showHelp(1)
	}
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
	//TODO: use environment variables
	host := "localhost"
	port := 5432
	user := "postgres"
	dbname := "test"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"dbname=%s sslmode=disable",
		host, port, user, dbname)
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

func main() {
	fetch := flag.Bool("fetch", false, "do fetch")
	read := flag.Bool("read", false, "do read")
	help := flag.Bool("help", false, "show help")
	file := flag.String("file", "", "file to fetch to or read from")
	flag.Parse()

	if *help {
		showHelp(0)
	}
	handleWrongInput(*fetch, *read, *file)

	switch {
	case *fetch:
		handleFetch(*file)
	case *read:
		handleRead(*file)
	default:
		fmt.Fprintf(os.Stderr, "Handle wrong input does not work!\n")
		showHelp(1)
	}
}
