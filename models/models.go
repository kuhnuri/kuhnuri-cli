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

type Project struct {
	// XXX this should actually be used by CDK and we just ignore it
	Publications []Publication `json:"publications"`
	Contexts     []Context     `json:"contexts"`
	Deliverables []Deliverable `json:"deliverables"`
}

type Publication struct {
	Transtype string  `json:"transtype"`
	Id        string  `json:"id"`
	Idref     string  `json:"idref"`
	Params    []Param `json:"params"`
}

type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Href  string `json:"href"`
}

type Context struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Idref string `json:"idref"`
	Input string `json:"input"`
}

type Deliverable struct {
	Name        string      `json:"name"`
	Id          string      `json:"id"`
	Context     Context     `json:"context"`
	Output      string      `json:"output"`
	Publication Publication `json:"publication"`
}

func NewContext(in string) *Context {
	return &Context{"", "", "", in}
}

func NewPublication(transtype string) *Publication {
	return &Publication{transtype, "", "", []Param{}}
}

func NewDeliverable(in string, transtype string, out string) *Deliverable {
	return &Deliverable{"", "", *NewContext(in), out, *NewPublication(transtype)}
}
