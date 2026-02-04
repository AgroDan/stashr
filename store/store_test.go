package store

import (
	"sort"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	s := New()
	defer s.Stop()

	s.Set("foo", "bar", 0)
	val, ok := s.Get("foo")
	if !ok || val != "bar" {
		t.Fatalf("expected (bar, true), got (%s, %v)", val, ok)
	}
}

func TestGetMissing(t *testing.T) {
	s := New()
	defer s.Stop()

	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected not found for missing key")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	defer s.Stop()

	s.Set("foo", "bar", 0)
	deleted := s.Delete("foo")
	if !deleted {
		t.Fatal("expected delete to return true")
	}

	_, ok := s.Get("foo")
	if ok {
		t.Fatal("expected key to be gone after delete")
	}
}

func TestDeleteMissing(t *testing.T) {
	s := New()
	defer s.Stop()

	deleted := s.Delete("nope")
	if deleted {
		t.Fatal("expected delete of missing key to return false")
	}
}

func TestList(t *testing.T) {
	s := New()
	defer s.Stop()

	s.Set("a", "1", 0)
	s.Set("b", "2", 0)
	s.Set("c", "3", 0)

	keys := s.List()
	sort.Strings(keys)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("unexpected keys: %v", keys)
	}
}

func TestTTLExpiry(t *testing.T) {
	s := New()
	defer s.Stop()

	s.Set("temp", "value", 50*time.Millisecond)

	val, ok := s.Get("temp")
	if !ok || val != "value" {
		t.Fatal("key should exist before TTL expires")
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = s.Get("temp")
	if ok {
		t.Fatal("key should have expired")
	}
}

func TestOverwrite(t *testing.T) {
	s := New()
	defer s.Stop()

	s.Set("key", "v1", 0)
	s.Set("key", "v2", 0)

	val, ok := s.Get("key")
	if !ok || val != "v2" {
		t.Fatalf("expected v2, got %s", val)
	}
}

func TestListExcludesExpired(t *testing.T) {
	s := New()
	defer s.Stop()

	s.Set("persist", "yes", 0)
	s.Set("temp", "no", 50*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	keys := s.List()
	if len(keys) != 1 || keys[0] != "persist" {
		t.Fatalf("expected only [persist], got %v", keys)
	}
}
