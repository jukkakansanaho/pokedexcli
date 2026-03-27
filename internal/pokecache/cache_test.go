package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	c := NewCache(1 * time.Hour)

	want := []byte("hello-world")
	c.Add("k1", want)
	got, ok := c.Get("k1")
	if !ok {
		t.Fatal("Get: want hit")
	}
	if string(got) != string(want) {
		t.Errorf("Get = %q; want %q", got, want)
	}
	// Mutating returned slice must not change cache.
	got[0] = 'X'
	got2, _ := c.Get("k1")
	if got2[0] != want[0] {
		t.Errorf("after mutating Get copy, cache corrupted: %q", got2)
	}
}

func TestGetMiss(t *testing.T) {
	c := NewCache(1 * time.Hour)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("Get missing key: want miss")
	}
}

func TestReapRemovesStaleEntries(t *testing.T) {
	interval := 20 * time.Millisecond
	c := NewCache(interval)
	c.Add("old", []byte("data"))
	time.Sleep(interval*2 + 15*time.Millisecond)
	_, ok := c.Get("old")
	if ok {
		t.Fatal("after reap interval, stale entry should be gone")
	}
}

func TestReapKeepsFreshEntries(t *testing.T) {
	interval := 50 * time.Millisecond
	c := NewCache(interval)
	c.Add("fresh", []byte("ok"))
	time.Sleep(interval / 2)
	got, ok := c.Get("fresh")
	if !ok || string(got) != "ok" {
		t.Fatalf("fresh entry missing: ok=%v got=%q", ok, got)
	}
}
