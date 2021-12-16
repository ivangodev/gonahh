package webapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vanyaio/gohh/data"
	"log"
	"net/http"
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

type respMsg struct {
	Msg string `json:"message"`
}

var requestsCache map[string][]byte

func init() {
	/* TODO: use real cache systems */
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

func allowCors(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func handleUnsetJobname(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)

	msg := respMsg{Msg: "jobname is not specified"}
	js, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal %s: %w\n", msg, err)
	} else {
		writeJson(w, js)
	}
}

func handleInternalError(w http.ResponseWriter, r *http.Request, e error) {
	w.WriteHeader(http.StatusInternalServerError)

	msg := respMsg{Msg: e.Error()}
	js, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal %s: %w\n", msg, err)
	} else {
		writeJson(w, js)
	}
}

func writeJson(w http.ResponseWriter, js []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getJobsNumber(db *sql.DB, jobName string) (int, error) {
	query := `SELECT COUNT(*) as jobsNumber FROM name WHERE name LIKE $1 `
	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	rows, err := db.Query(query, likeJobName)
	if err != nil {
		return 0, err
	}

	var jobsNumber int
	for rows.Next() {
		err = rows.Scan(&jobsNumber)
		if err != nil {
			return 0, err
		}
	}
	return jobsNumber, nil
}

func createOrderedRate(db *sql.DB, jobName string) error {
	//This is better to be view, but it seems it doesn't support placeholders
	query := `CREATE TEMPORARY TABLE ordered_rate (
		word VARCHAR ( 255 ) NOT NULL,
		rate integer
	);`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	likeJobName := fmt.Sprintf("%%%v%%", jobName)
	query = `INSERT INTO ordered_rate SELECT engwords.word as keyword,
		ROUND(COUNT(t.job_id)*100.0 /
		(SELECT COUNT(*) FROM name WHERE name LIKE $1 ), 0) as rate
		FROM engwords INNER JOIN
		( SELECT * FROM name WHERE name LIKE $2 ) as t
		ON t.job_id=engwords.job_id WHERE word <> $3
		GROUP BY word ORDER BY rate DESC LIMIT 30;`
	_, err = db.Exec(query, likeJobName, likeJobName, jobName)
	if err != nil {
		return err
	}

	return nil
}

func fillAllTop(db *sql.DB, resp *DataJSON) error {
	query := "SELECT * FROM ordered_rate ;"
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	resp.AllTop = make([]KeywordRate, 0)
	for rows.Next() {
		var kr KeywordRate
		err = rows.Scan(&kr.Keyword, &kr.Rate)
		if err != nil {
			return err
		}
		resp.AllTop = append(resp.AllTop, kr)
	}

	return nil
}

func fillCategoriesTop(db *sql.DB, resp *DataJSON) error {
	query := `CREATE VIEW full_category AS
	SELECT ordered_rate.word, category
	FROM ordered_rate LEFT JOIN category on ordered_rate.word=category.word;`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	query = `SELECT ordered_rate.word, rate, category
			 FROM ordered_rate JOIN full_category
			 ON ordered_rate.word=full_category.word
			 ORDER BY category, rate DESC;`
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	resp.CategoriesTop = make([]CategoryTop, 0)
	currTop := new(CategoryTop)
	for rows.Next() {
		var kr KeywordRate
		var category string
		err = rows.Scan(&kr.Keyword, &kr.Rate, &category)
		if err != nil {
			var null_category sql.NullString
			err = rows.Scan(&kr.Keyword, &kr.Rate, &null_category)
			if err != nil {
				return err
			}
			category = "Other"
		}

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
	return nil
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	allowCors(w, r)

	jobName := strings.ToLower(r.FormValue("jobname"))
	if jobName == "" {
		handleUnsetJobname(w, r)
		return
	}

	cachedResp, toCache := requestsCache[jobName]
	if toCache && cachedResp != nil {
		writeJson(w, cachedResp)
		return
	}

	db, err := data.OpenDB()
	if err != nil {
		handleInternalError(w, r, fmt.Errorf("Failed to open DB: %w", err))
		return
	}
	defer db.Close()

	var resp DataJSON
	resp.JobName = jobName
	resp.JobsNumber, err = getJobsNumber(db, jobName)
	if err != nil {
		handleInternalError(w, r, fmt.Errorf("Failed get jobs number: %w", err))
		return
	}

	err = createOrderedRate(db, jobName)
	if err != nil {
		handleInternalError(w, r,
			fmt.Errorf("Failed to create ordered rate: %w", err))
		return
	}

	err = fillAllTop(db, &resp)
	if err != nil {
		handleInternalError(w, r,
			fmt.Errorf("Failed to fill all top: %w", err))
		return
	}

	err = fillCategoriesTop(db, &resp)
	if err != nil {
		handleInternalError(w, r,
			fmt.Errorf("Failed to fill categories top: %w", err))
		return
	}

	js, err := json.Marshal(resp)
	if err != nil {
		handleInternalError(w, r,
			fmt.Errorf("Failed to marshal %v:  %w", resp, err))
		return
	}

	writeJson(w, js)
	if toCache {
		requestsCache[jobName] = js
	}
}

func Main() {
	http.HandleFunc("/api/", searchHandler)
	fs := http.FileServer(http.Dir("./webapp/static"))
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
