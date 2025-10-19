package models

import (
	"testing"
	"time"
)

func TestModelStructs(t *testing.T) {
	now := time.Now()
	j := Jar{ID: "1", Name: "test", CreatedAt: now}
	if j.ID != "1" || j.Name != "test" {
		t.Fatalf("unexpected jar values: %+v", j)
	}

	r := Request{ID: "r1", CreatedAt: now, Method: "GET", Path: "/"}
	if r.Method != "GET" || r.Path != "/" {
		t.Fatalf("unexpected request values: %+v", r)
	}
}
