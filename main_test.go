package main

import (
	"testing"
	"time"
)

func TestSinceAndBefore(t *testing.T) {
	now, _ := time.Parse("2006-01-02 15:04:05", "1979-01-01 07:23:00")
	since, before := sinceAndBefore(now)

	if since != "1979-01-01T06:00:00Z" {
		t.Fatalf("failed test: %s", since)
	}

	if before != "1979-01-01T07:23:00Z" {
		t.Fatalf("failed test: %s", before)
	}
}

func TestIsMention(t *testing.T) {
	mention := Notification{Reason: "mention"}
	notMention := Notification{Reason: "any"}

	if !mention.IsMention() {
		t.Fatalf("failed test")
	}

	if notMention.IsMention() {
		t.Fatalf("failed test")
	}
}
