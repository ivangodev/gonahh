package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vanyaio/gohh/data"
	"github.com/vanyaio/gohh/fetcher"
	"github.com/vanyaio/gohh/vacancy"
	"github.com/vanyaio/gohh/webapp"
	"os"
)

type category struct {
	name  string
	words []string
}

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
	go fetcher.FetchRateLimit(1000, 7) //https://github.com/hhru/api/issues/74
	fetcher.FetchAndLogVacs(filename)
	close(fetcher.FetchQueue)
}

func checkErr(e error) {
	if e != nil {
		panic(e)
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
	db := data.Db
	file, err := os.Open(filename)
	checkErr(err)
	defer file.Close()

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
	vacs, err := vacancy.ReadVacancies(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to read vacancies: %s", err))
	}

	err = vacancy.DumpVacaniesToDB(vacs)
	if err != nil {
		panic(fmt.Sprintf("Failed to dump vacancies DB: %s", err))
	}
}

func setCategory(c *category) {
	db := data.Db

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
