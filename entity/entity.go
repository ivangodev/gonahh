package entity

type KeywordRate struct {
	Keyword string `json:"Keyword"`
	Rate    int    `json:"Rate"`
}

type CategoryTop struct {
	Category string        `json:"Category"`
	Top      []KeywordRate `json:"Top"`
}

type ServiceResp struct {
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

type DescrName struct {
	Descr string
	Name  string
}
type URLtoDescrName map[string]DescrName

type JobInfo struct {
	Name     string
	Keywords []string
}
type Schema struct {
	URLtoJobInfo    map[string]JobInfo
	KeywordCategory map[string]string
}
