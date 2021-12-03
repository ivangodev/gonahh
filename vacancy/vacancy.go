package vacancy

import (
	"bytes"
	"encoding/gob"
	"github.com/vanyaio/gohh/data"
	"io/ioutil"
	"strings"
)

type Vacancy struct {
	Url      string
	Name     string
	EngWords []string
}

func storeToGOB(data interface{}, filename string) error {
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(data)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, buffer.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func loadFromGOB(data interface{}, filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buffer)
	err = dec.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func LogVacancies(filename string, vacs []Vacancy) error {
	return storeToGOB(vacs, filename)
}

func ReadVacancies(filename string) ([]Vacancy, error) {
	var res []Vacancy
	err := loadFromGOB(&res, filename)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *Vacancy) dumpToDB() error {
	db := data.Db
	stat := `INSERT INTO url(url) VALUES ($1)`
	_, err := db.Exec(stat, v.Url)
	if err != nil {
		return nil /* It's likely violates duplicates - drop such job. */
	}

	stat = `SELECT job_id FROM url WHERE url = $1`
	rows, err := db.Query(stat, v.Url)
	if err != nil {
		return err
	}

	var job_id int
	for rows.Next() {
		err = rows.Scan(&job_id)
		if err != nil {
			return err
		}
	}

	stat = `INSERT INTO name(job_id, name) VALUES ($1, $2)`
	v.Name = strings.ToLower(v.Name)
	_, err = db.Exec(stat, job_id, v.Name)
	if err != nil {
		return err
	}

	for _, w := range v.EngWords {
		stat = `INSERT INTO engwords(job_id, word) VALUES ($1, $2)`
		_, err = db.Exec(stat, job_id, w)
		if err != nil {
			return err
		}
	}

	return nil
}

func DumpVacaniesToDB(vacs []Vacancy) error {
	for _, v := range vacs {
		err := v.dumpToDB()
		if err != nil {
			return err
		}
	}
	return nil
}
