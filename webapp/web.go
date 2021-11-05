package webapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strings"
)

type KeywordRate struct {
	Keyword string `json:"Keyword"`
	Rate    int    `json:"Rate"`
}

type CategoryTop struct {
	Category string        `json:"Category"`
	Top      []KeywordRate `json:"Top"`
}

type DataJSON struct {
	JobName       string        `json:"JobName"`
	JobsNumber    int           `json:"JobsNumber"`
	AllTop        []KeywordRate `json:"AllTop"`
	CategoriesTop []CategoryTop `json:"CategoriesTop"`
}

type Keyword struct {
	Name     string `json:"keyword"`
	Rate     int    `json:"rate"`
	Category string `json:"category"`
}

var requestsCache map[string][]byte

func initCache() {
	/* It's awful */
	requestsCache = make(map[string][]byte)
	requestsCache["go"] = nil
	requestsCache["golang"] = nil
	requestsCache["ruby"] = nil
	requestsCache["swift"] = nil
	requestsCache["kotlin"] = nil
	requestsCache["java"] = nil
	requestsCache["javascript"] = nil
	requestsCache["js"] = nil
	requestsCache["c#"] = nil
	requestsCache[".net"] = nil
	requestsCache["python"] = nil
	requestsCache["php"] = nil
	requestsCache["front-end"] = nil
	requestsCache["back-end"] = nil
	requestsCache["backend"] = nil
	requestsCache["frontend"] = nil
	requestsCache["ios"] = nil
	requestsCache["embedded"] = nil
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

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	jobName := strings.ToLower(r.FormValue("jobname"))

	cachedResp, toCache := requestsCache[jobName]
	if toCache && cachedResp != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cachedResp)
		return
	}

	db := openDB()
	defer db.Close()

	var resp DataJSON
	resp.JobName = jobName

	//Get jobs Number
	query := `SELECT COUNT(*) as jobsNumber FROM name WHERE name LIKE $1 `
	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	rows, err := db.Query(query, likeJobName)
	checkErr(err)
	var jobsNumber int
	for rows.Next() {
		err = rows.Scan(&jobsNumber)
		checkErr(err)
	}
	resp.JobsNumber = jobsNumber

	//Get Top without category
	query = `CREATE TABLE ordered_rate (
		word VARCHAR ( 255 ) NOT NULL,
		rate integer
	);`
	_, err = db.Exec(query)
	checkErr(err)

	query = `INSERT INTO ordered_rate SELECT engwords.word as keyword,
		ROUND(COUNT(t.job_id)*100.0 /
		(SELECT COUNT(*) FROM name WHERE name LIKE $1 ), 0) as rate
		FROM engwords INNER JOIN
		( SELECT * FROM name WHERE name LIKE $2 ) as t
		ON t.job_id=engwords.job_id WHERE word <> $3
		GROUP BY word ORDER BY rate DESC LIMIT 30;`
	_, err = db.Exec(query, likeJobName, likeJobName, jobName)
	checkErr(err)

	query = "SELECT * FROM ordered_rate ;"
	rows, err = db.Query(query)
	checkErr(err)
	resp.AllTop = make([]KeywordRate, 0)
	for rows.Next() {
		var kr KeywordRate
		err = rows.Scan(&kr.Keyword, &kr.Rate)
		checkErr(err)
		resp.AllTop = append(resp.AllTop, kr)
	}

	//Move keywords  missing own category to "other" category.
	query = `CREATE TABLE full_category (
		word VARCHAR ( 255 ) NOT NULL UNIQUE,
		category VARCHAR ( 255 )
	);`
	_, err = db.Exec(query)
	checkErr(err)

	query = `INSERT INTO full_category
	SELECT ordered_rate.word, category
	FROM ordered_rate LEFT JOIN category on ordered_rate.word=category.word;`
	_, err = db.Exec(query)
	checkErr(err)

	query = "UPDATE full_category SET category='Прочее' where category is NULL;"
	_, err = db.Exec(query)
	checkErr(err)

	//Get top across categories
	query = `SELECT ordered_rate.word, rate, category
			 FROM ordered_rate JOIN full_category
			 ON ordered_rate.word=full_category.word
			 ORDER BY category, rate DESC;`
	rows, err = db.Query(query)
	checkErr(err)
	resp.CategoriesTop = make([]CategoryTop, 0)
	currTop := new(CategoryTop)
	for rows.Next() {
		var kr KeywordRate
		var category string
		err = rows.Scan(&kr.Keyword, &kr.Rate, &category)
		checkErr(err)

		if category != currTop.Category {
			if currTop.Category != "" {
				resp.CategoriesTop = append(resp.CategoriesTop, *currTop)
			}
			currTop = new(CategoryTop)
			currTop.Category = category
		}
		currTop.Top = append(currTop.Top, kr)
	}
	resp.CategoriesTop = append(resp.CategoriesTop, *currTop)

	//Cleanup.
	query = `DROP TABLE ordered_rate cascade`
	_, err = db.Exec(query)
	checkErr(err)
	query = `DROP TABLE full_category cascade`
	_, err = db.Exec(query)
	checkErr(err)

	js, err := json.Marshal(resp)
	checkErr(err)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	if toCache {
		requestsCache[jobName] = js
	}
}

func Main() {
	initCache()
	http.HandleFunc("/api/", searchHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
