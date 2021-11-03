package main

import (
	"github.com/vanyaio/gohh/fetcher"
	"fmt"
	"os"
	"flag"
)

func showHelp(error_code int) {
	fmt.Println("usage: main --file /path/to/file [--read|--fetch]")
	fmt.Println("Fetch to or read from file vacancies info")
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

func handleRead(filename string) {
	fmt.Printf("Read operation currently is unsupported (%s)\n", filename)
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
