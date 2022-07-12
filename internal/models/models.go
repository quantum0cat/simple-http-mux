package models

type UrlsDto struct {
	Urls []string `json:"urls"`
}

type Response struct {
	Url      string
	Response string
}
type ResponsesDto struct {
	_ []Response
}
