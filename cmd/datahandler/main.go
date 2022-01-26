package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/ivangodev/fefa/pkg/fefa"
	"github.com/ivangodev/gonahh/entity"
	"github.com/ivangodev/gonahh/fetcher"
	"github.com/ivangodev/gonahh/fetcher/hh"
	"github.com/ivangodev/gonahh/repository"
	"github.com/ivangodev/gonahh/repository/psql"
	"os"
	"strings"
)

type helpFiles struct {
	fetch  string
	ban    string
	categs string
}

func handleFetch(files helpFiles, repo repository.RepoI) error {
	//Rere limit according to https://github.com/hhru/api/issues/74
	opts := fefa.RateLimitOpts{Interval: 1000, ReqsRate: 7}
	fefa.FeFa(fetcher.NewRoot(hh.Callbacks), &opts)

	schema := entity.Schema{URLtoJobInfo: make(map[string]entity.JobInfo),
		KeywordCategory: make(map[string]string)}
	var ban *banner
	if files.ban != "" {
		f, err := os.Open(files.ban)
		if err != nil {
			return fmt.Errorf("Failed to open %s: %w", files.ban, err)
		}
		ban = newKeywordsBanner(f)
	}
	for url, descrName := range fetcher.ResVacsInfo {
		name := replaceSynonym(strings.ToLower(descrName.Name))
		keywords := extractEngWords(descrName.Descr)

		if files.ban != "" {
			keywords = ban.banKeywords(keywords)
		}
		schema.URLtoJobInfo[url] = entity.JobInfo{Name: name, Keywords: keywords}
	}

	if files.categs != "" {
		f, err := os.Open(files.categs)
		if err != nil {
			return fmt.Errorf("Failed to open %s: %w", files.categs, err)
		}
		schema.KeywordCategory = categorize(f, schema.URLtoJobInfo)
	}

	f, err := os.OpenFile(files.fetch, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %w", files.fetch, err)
	}
	err = repo.Write(schema, f)
	if err != nil {
		return fmt.Errorf("Failed to write to file: %w", err)
	}

	return nil
}

func handleRead(file string, repo repository.RepoI) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("Failed to open %s: %w", file, err)
	}
	if err := repo.Read(f); err != nil {
		return fmt.Errorf("Failed to read %s: %w", file, err)
	}
	return nil
}

func main() {
	mockHH := flag.Bool("mockhh", false, "Start mock HH server")
	mockDB := flag.Bool("mockdb", false, "Use mock database")
	fetchTo := flag.String("fetchto", "", "File to fetch to")
	ban := flag.String("banwords", "", "File with banned words")
	categs := flag.String("categories", "", "File with categories")
	readFrom := flag.String("readfrom", "", "File to read from")
	flag.Parse()
	var db *sql.DB

	if *mockHH {
		mockServer()
		hh.ApiUrl = "http://localhost:8080/pageinfo"
	}

	if *mockDB {
		var purge func()
		db, purge = initMockDB()
		defer purge()
	} else {
		var err error
		db, err = psql.OpenDB()
		if err != nil {
			panic(fmt.Sprintf("Failed to open DB: %s", err))
		}
	}

	repo := psql.NewPSql(db)
	if *fetchTo != "" {
		files := helpFiles{fetch: *fetchTo, ban: *ban, categs: *categs}
		if err := handleFetch(files, repo); err != nil {
			panic(fmt.Sprintf("Failed to fetch: %s", err))
		}
	}
	if *readFrom != "" {
		if err := handleRead(*readFrom, repo); err != nil {
			panic(fmt.Sprintf("Failed to read: %s", err))
		}
	}
}
