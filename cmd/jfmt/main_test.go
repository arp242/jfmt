package main

import (
	"strings"
	"testing"
)

func TestUsage(t *testing.T) {
	if strings.Contains(usage, "\t") {
		t.Error("usage contains tabs")
	}
}
