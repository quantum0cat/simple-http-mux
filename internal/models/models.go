package models

import "encoding/json"

type UrlsDto struct {
	Urls []string `json:"urls"`
}

func (u *UrlsDto) Marshal() []byte {
	data, err := json.Marshal(*u)
	if err != nil {
		return []byte{}
	}
	return data
}

type Response struct {
	Url      string `json:"url"`
	Response string `json:"response"`
}
