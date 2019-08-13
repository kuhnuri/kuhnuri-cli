package models

import "net/url"

type Create struct {
	Transtype *[]string `json:"transtype"`
	Input     *string   `json:"input"`
	Output    *string   `json:"output"`
}

func NewCreate(transtype []string, input *url.URL, output *url.URL) *Create {
	var o string
	if output == nil {
		o = output.String()
	}
	i := input.String()
	return &Create{&transtype, &i, &o}
}

type Upload struct {
	Url    string `json:"url"`
	Upload string `json:"upload"`
}

type Job struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	Output string `json:"output"`
}
