package mock

import (
	"fmt"
	"github.com/ivangodev/gonahh/entity"
	"io"
)

type repoMockData struct {
	word     string
	rate     int
	category string
}

type repoMock struct {
	data []repoMockData
}

const (
	CorrectJobName = "correctjob"
	jobsNumber     = 10
)

func NewRepoMock() *repoMock {
	data := []repoMockData{{word: "Java",
		rate:     50,
		category: "Language"},
		{word: "C++",
			rate:     32,
			category: "Language"},
		{word: "Gin",
			rate:     21,
			category: "Framework"},
		{word: "MySQL",
			rate:     26,
			category: "Database"},
	}
	return &repoMock{data: data}
}

func (r *repoMock) GetJobsNumber(jobName string) (int, error) {
	if jobName != CorrectJobName {
		return 0, nil
	}
	return jobsNumber, nil
}

func (r *repoMock) GetAllTop(jobName string) ([]entity.KeywordRate, error) {
	if jobName != CorrectJobName {
		return make([]entity.KeywordRate, 0), nil
	}
	return []entity.KeywordRate{{Keyword: "Java",
		Rate: 50},
		{Keyword: "C++",
			Rate: 32},
		{Keyword: "MySQL",
			Rate: 26},
		{Keyword: "Gin",
			Rate: 21},
	}, nil
}

func (r *repoMock) GetCategories(jobName string) ([]string, error) {
	if jobName != CorrectJobName {
		return make([]string, 0), nil
	}
	return []string{"Language", "Framework", "Database"}, nil
}

func (r *repoMock) GetCategoryTop(jobName, category string) (entity.CategoryTop, error) {
	if jobName != CorrectJobName {
		return entity.CategoryTop{Category: category, Top: []entity.KeywordRate{}}, nil
	}
	switch category {
	case "Language":
		return entity.CategoryTop{Category: category,
			Top: []entity.KeywordRate{
				{"Java", 50},
				{"C++", 32},
			}}, nil
	case "Framework":
		return entity.CategoryTop{Category: category,
			Top: []entity.KeywordRate{{"Gin", 21}}}, nil
	case "Database":
		return entity.CategoryTop{Category: category,
			Top: []entity.KeywordRate{{"MySQL", 26}}}, nil
	default:
		return entity.CategoryTop{Category: category,
			Top: []entity.KeywordRate{}}, nil
	}
}

//Useful for testing. Not a part of Repository interface.
func (r *repoMock) GetServiceResp(jobName string) (resp entity.ServiceResp) {
	resp.JobName = jobName
	resp.JobsNumber, _ = r.GetJobsNumber(jobName)
	resp.AllTop = []entity.KeywordRate{}
	resp.CategoriesTop = []entity.CategoryTop{}

	if jobName == CorrectJobName {
		resp.AllTop = []entity.KeywordRate{{"Java", 50},
			{"C++", 32},
			{"MySQL", 26},
			{"Gin", 21},
		}

		resp.CategoriesTop = []entity.CategoryTop{
			{
				Category: "Language",
				Top: []entity.KeywordRate{
					{"Java", 50},
					{"C++", 32}},
			},
			{
				Category: "Framework",
				Top:      []entity.KeywordRate{{"Gin", 21}},
			},
			{
				Category: "Database",
				Top:      []entity.KeywordRate{{"MySQL", 26}},
			},
		}
	}

	return
}

func (repo *repoMock) Write(data entity.Schema, w io.Writer) error {
	return fmt.Errorf("'Write' is unsupported")
}

func (repo *repoMock) Read(r io.Reader) error {
	return fmt.Errorf("'Read' is unsupported")
}
