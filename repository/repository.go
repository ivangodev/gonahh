package repository

import (
	"github.com/ivangodev/gonahh/entity"
	"io"
)

type RepoI interface {
	GetJobsNumber(jobName string) (int, error)
	GetAllTop(jobName string) ([]entity.KeywordRate, error)
	GetCategories(jobName string) ([]string, error)
	GetCategoryTop(jobName, category string) (entity.CategoryTop, error)
	Write(data entity.Schema, w io.Writer) error
	Read(r io.Reader) error
}
