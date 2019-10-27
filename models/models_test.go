package models

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestParseProject(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/project.json")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	project := Project{}
	err = json.Unmarshal(buf, &project)
	if err != nil {
		t.Fatalf("Failed to read JSON: %v", err)
	}
	if len(project.Deliverables) != 2 {
		t.Fatalf("Incorrect deliverables count: %d", len(project.Deliverables))
	}
}
