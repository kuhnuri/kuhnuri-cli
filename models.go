package main

type Create struct {
	Transtype *string `json:"transtype"`
	Input     *string `json:"input"`
	Output    *string `json:"output"`
}

func NewCreate(transtype string, input string, output string) *Create {
	var o *string
	if len(output) == 0 {
		o = &output
	}
	return &Create{&transtype, &input, o}
}

type Upload struct {
	Url    string `json:"url"`
	Upload string `json:"upload"`
}

type Job struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}
