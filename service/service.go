package service

import (
	"fmt"
	"github.com/ivangodev/gonahh/entity"
	"github.com/ivangodev/gonahh/repository"
	"strings"
)

type ServiceI interface {
	HandleReq(jobName string) (entity.ServiceResp, error)
}

type Service struct {
	repo repository.RepoI
}

func NewService(r repository.RepoI) *Service {
	return &Service{repo: r}
}

func (s *Service) getCategoriesTop(jobName string) ([]entity.CategoryTop, error) {
	tops := make([]entity.CategoryTop, 0)
	categories, err := s.repo.GetCategories(jobName)
	if err != nil {
		err = fmt.Errorf("Failed to get categories: %w", err)
		return nil, err
	}

	for _, c := range categories {
		top, err := s.repo.GetCategoryTop(jobName, c)
		if err != nil {
			err = fmt.Errorf("Failed to get top for category %v: %w", c, err)
			return nil, err
		}
		tops = append(tops, top)
	}

	return tops, nil
}

func truncate(t []entity.KeywordRate, thresholdRate int) []entity.KeywordRate {
	var i int
	for ; i < len(t) && t[i].Rate >= thresholdRate; i++ {
	}
	//Don't make the top empty
	if i == 0 {
		i = 1
	}
	return t[:i]
}

func upJob(t []entity.KeywordRate, jobName string) []entity.KeywordRate {
	for i, kr := range t {
		if kr.Keyword == jobName {
			kr.Rate = 100
			t = append(t[:i], t[i+1:]...)
			t = append([]entity.KeywordRate{kr}, t...)
			break
		}
	}
	return t
}

func replaceSynonym(word string) string {
	//FIXME: this is copy-paste from extractor.go
	switch word {
	case "js":
		return "javascript"
	case "ror":
		return "rails"
	case "vue.js":
		return "vue"
	case "golang":
		return "go"
	case "postgres":
		return "postgresql"
	default:
		return word
	}
}

func cleanJobName(jobName string) string {
	return replaceSynonym(strings.TrimSpace(strings.ToLower(jobName)))
}

func (s *Service) HandleReq(jobName string) (resp entity.ServiceResp, err error) {
	jobName = cleanJobName(jobName)
	resp.JobName = jobName
	resp.JobsNumber, err = s.repo.GetJobsNumber(jobName)
	if err != nil {
		err = fmt.Errorf("Failed to get jobs number: %w", err)
		return resp, err
	}

	resp.AllTop, err = s.repo.GetAllTop(jobName)
	if err != nil {
		err = fmt.Errorf("Failed to get all top: %w", err)
		return resp, err
	}

	resp.CategoriesTop, err = s.getCategoriesTop(jobName)
	if err != nil {
		err = fmt.Errorf("Failed to get categories top: %w", err)
		return resp, err
	}

	thresholdRate := 10
	resp.AllTop = truncate(resp.AllTop, thresholdRate)
	resp.AllTop = upJob(resp.AllTop, jobName)
	for i, t := range resp.CategoriesTop {
		resp.CategoriesTop[i].Top = truncate(t.Top, thresholdRate)
		resp.CategoriesTop[i].Top = upJob(resp.CategoriesTop[i].Top, jobName)
	}

	return
}
