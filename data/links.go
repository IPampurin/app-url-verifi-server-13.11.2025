package data

type LinkStatus string

const (
	StatusAvailable    LinkStatus = "available"
	StatusNotAvailable LinkStatus = "not available"
)

type Request struct {
	Links []string `json:"links"`
}

type Response struct {
	Links    map[string]LinkStatus `json:"links"`
	LinksNum int                   `json:"links_num"`
}
