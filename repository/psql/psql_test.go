package psql

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/ivangodev/gonahh/entity"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

func fillDB(t *testing.T) {
	q := `
    INSERT INTO url values (1, 'gourl1');
    INSERT INTO url values (2, 'gourl2');
    INSERT INTO url values (3, 'gourl3');
    INSERT INTO url values (4, 'gourl4');
    INSERT INTO url values (5, 'gourl5');
    INSERT INTO url values (6, 'jsurl6');

    INSERT INTO name(job_id, name) values (1, 'Go developer');
    INSERT INTO name(job_id, name) values (2, 'Go developer');
    INSERT INTO name(job_id, name) values (3, 'Go developer');
    INSERT INTO name(job_id, name) values (4, 'Go developer');
    INSERT INTO name(job_id, name) values (5, 'Go developer');

    INSERT INTO engwords(job_id, word) values (1, 'Go');
    INSERT INTO engwords(job_id, word) values (2, 'Go');
    INSERT INTO engwords(job_id, word) values (3, 'Go');
    INSERT INTO engwords(job_id, word) values (4, 'Go');
    INSERT INTO engwords(job_id, word) values (5, 'Go');
    INSERT INTO engwords(job_id, word) values (1, 'Git');
    INSERT INTO engwords(job_id, word) values (2, 'Git');
    INSERT INTO engwords(job_id, word) values (3, 'Git');
    INSERT INTO engwords(job_id, word) values (4, 'Git');
    INSERT INTO engwords(job_id, word) values (5, 'Git');
    INSERT INTO engwords(job_id, word) values (1, 'MySQL');
    INSERT INTO engwords(job_id, word) values (2, 'MySQL');
    INSERT INTO engwords(job_id, word) values (3, 'MySQL');
    INSERT INTO engwords(job_id, word) values (4, 'MySQL');
    INSERT INTO engwords(job_id, word) values (1, 'PHP');
    INSERT INTO engwords(job_id, word) values (2, 'PHP');

    INSERT INTO name(job_id, name) values (6, 'Frontend developer');
    INSERT INTO engwords(job_id, word) values (6, 'JavaScript');

    INSERT INTO category(word, category) values ('Go', 'Language');
    INSERT INTO category(word, category) values ('PHP', 'Language');
    INSERT INTO category(word, category) values ('JavaScript', 'Language');
    INSERT INTO category(word, category) values ('MySQL', 'Database');
    INSERT INTO category(word, category) values ('Git', '');
    `
	_, err := db.Exec(q)
	if err != nil {
		t.Fatalf("Failed to populate tables: %s", err)
	}
}

var errUnsupportedJobName = fmt.Errorf("Unsupported job")
var errUnknownCategory = fmt.Errorf("Unknown category")
var supportedJobName = "Go"

func expectedJobsNumber(jobName string) (int, error) {
	if jobName != supportedJobName {
		return 0, errUnsupportedJobName
	}
	return 5, nil
}

func expectedAllTop(jobName string) ([]entity.KeywordRate, error) {
	if jobName != supportedJobName {
		return nil, errUnsupportedJobName
	}
	return []entity.KeywordRate{{Keyword: "Go",
		Rate: 100},
		{Keyword: "Git",
			Rate: 100},
		{Keyword: "MySQL",
			Rate: 80},
		{Keyword: "PHP",
			Rate: 40},
	}, nil
}

func expectedCategories(jobName string) ([]string, error) {
	if jobName != supportedJobName {
		return nil, errUnsupportedJobName
	}
	return []string{"Language", "Database", ""}, nil
}

func expectedCategoryTop(jobName, category string) (entity.CategoryTop, error) {
	if jobName != supportedJobName {
		return entity.CategoryTop{}, errUnsupportedJobName
	}

	switch category {
	case "Language":
		return entity.CategoryTop{Category: category,
			Top: []entity.KeywordRate{
				{"Go", 100},
				{"PHP", 40},
			}}, nil
	case "Database":
		return entity.CategoryTop{Category: category,
			Top: []entity.KeywordRate{
				{"MySQL", 80},
			}}, nil
	case "":
		return entity.CategoryTop{Category: "",
			Top: []entity.KeywordRate{
				{"Git", 100},
			}}, nil
	default:
		return entity.CategoryTop{}, errUnknownCategory
	}
}

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=dbname",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/dbname?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestCallbacks(t *testing.T) {
	r := NewPSql(db)
	fillDB(t)
	jobName := "Go"

	actualJN, err := r.GetJobsNumber(jobName)
	if err != nil {
		t.Fatalf("Failed to get jobs number: %s", err)
	}
	expectedJN, err := expectedJobsNumber(jobName)
	if err != nil {
		t.Fatalf("Failed to get expected jobs number: %s", err)
	}
	if actualJN != expectedJN {
		t.Fatalf("Jobs number: actual %v VS expected %v", actualJN, expectedJN)
	}

	actualAT, err := r.GetAllTop(jobName)
	if err != nil {
		t.Fatalf("Failed to get all top: %s", err)
	}
	expectedAT, err := expectedAllTop(jobName)
	if err != nil {
		t.Fatalf("Failed to get expected all top: %s", err)
	}
	if !reflect.DeepEqual(actualAT, expectedAT) {
		t.Fatalf("All top: actual %v VS expected %v", actualAT, expectedAT)
	}

	actualC, err := r.GetCategories(jobName)
	if err != nil {
		t.Fatalf("Failed to get categories: %s", err)
	}
	expectedC, err := expectedCategories(jobName)
	if err != nil {
		t.Fatalf("Failed to get expected categories: %s", err)
	}
	sort.Strings(actualC)
	sort.Strings(expectedC)
	if !reflect.DeepEqual(actualC, expectedC) {
		t.Fatalf("Categories: actual %v VS expected %v", actualC, expectedC)
	}

	for _, c := range actualC {
		log.Printf("Test category %v\n", c)
		actualT, err := r.GetCategoryTop(jobName, c)
		if err != nil {
			t.Fatalf("Failed to category top: %s", err)
		}
		expectedT, err := r.GetCategoryTop(jobName, c)
		if err != nil {
			t.Fatalf("Failed to get expected category top: %s", err)
		}
		if !reflect.DeepEqual(actualT, expectedT) {
			t.Fatalf("Category top: actual %v VS expected %v", actualT, expectedT)
		}
		log.Printf("OK\n")
	}

	err = r.deleteTables()
	if err != nil {
		t.Fatalf("Failed to delete tables: %s", err)
	}
	r = NewPSql(db)
	data := entity.Schema{URLtoJobInfo: map[string]entity.JobInfo{
		"example.com/1": {"Java Dev",
			[]string{"Java", "Git"}},
		"example.com/2": {"Go Dev",
			[]string{"Go", "MySQL"}},
	},
		KeywordCategory: map[string]string{
			"Java":  "Language",
			"Go":    "Language",
			"MySQL": "Databases",
		},
	}
	var buf bytes.Buffer
	err = r.Write(data, &buf)
	if err != nil {
		t.Fatalf("Failed to write data: %s", err)
	}

	err = r.Read(&buf)
	if err != nil {
		t.Fatalf("Failed to read data: %s", err)
	}

	q := `SELECT name FROM name WHERE job_id = 1`
	rows, err := r.Db.Query(q)
	if err != nil {
		t.Fatalf("Failed to exec query: %s: %s", q, err)
	}
	if rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			t.Fatalf("Failed to read query result: %s: %s", q, err)
		}
		if name != "Java Dev" && name != "Go Dev" {
			t.Fatalf("Unexpected query %s result: name %s", q, name)
		}
	} else {
		t.Fatalf("Query is unexpectedly empty: %s", q)
	}

	q = `SELECT word FROM engwords WHERE job_id = 1`
	rows, err = r.Db.Query(q)
	if err != nil {
		t.Fatalf("Failed to exec query: %s: %s", q, err)
	}
	actual := make([]string, 0)
	for rows.Next() {
		var kw string
		err = rows.Scan(&kw)
		if err != nil {
			t.Fatalf("Failed to read query result: %s: %s", q, err)
		}
		actual = append(actual, kw)
	}
	want1 := []string{"Java", "Git"}
	want2 := []string{"Go", "MySQL"}
	sort.Strings(actual)
	sort.Strings(want1)
	sort.Strings(want2)
	if !reflect.DeepEqual(actual, want1) && !reflect.DeepEqual(actual, want2) {
		t.Fatalf("Unexpected keywords: actual %v VS want %v or %v",
			actual, want1, want2)
	}

	q = `SELECT word, category FROM category`
	rows, err = r.Db.Query(q)
	if err != nil {
		t.Fatalf("Failed to exec query: %s: %s", q, err)
	}
	type keywordCategory struct {
		keyword  string
		category string
	}
	actualKC := make([]keywordCategory, 0)
	for i := 0; rows.Next(); i++ {
		actualKC = append(actualKC, keywordCategory{})
		err = rows.Scan(&actualKC[i].keyword, &actualKC[i].category)
		if err != nil {
			t.Fatalf("Failed to read query result: %s: %s", q, err)
		}
	}
	if len(actualKC) != 3 {
		t.Fatalf("Unexpected categories for keywords: %v", actualKC)
	}
	for _, keywordCateg := range actualKC {
		k := keywordCateg.keyword
		c := keywordCateg.category
		if v, ok := data.KeywordCategory[k]; !ok || v != c {
			t.Fatalf("Unexpected categories for keywords: %v", actualKC)
		}
	}
}
