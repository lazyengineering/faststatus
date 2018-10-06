package server

import (
	"io/ioutil"
	"testing"
)

func TestBoltStore(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "_test")
	if err != nil {
		t.Fatalf("creating test file: %+v", err)
	}
	fnm := tmpfile.Name()
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("closing test file: %+v", err)
	}
	s, err := Current()
	if err != nil {
		t.Fatalf("creating new current server: %+v", err)
	}
	if s == nil {
		t.Fatalf("expected non-nil handler")
	}
	c, ok := s.(*current)
	if !ok {
		t.Fatalf("expected *current as handler")
	}
	if c.store != nil {
		t.Fatalf("expected nil store in new current")
	}
	if err := BoltStore(fnm)(c); err != nil {
		t.Fatalf("BoltStore(%s)(*current) = %+v, expected no errors", fnm, err)
	}
	if c.store == nil {
		t.Fatalf("expected non-nil store after calling BoltStore option")
	}
	if err := BoltStore("")(c); err == nil {
		t.Fatalf("BoltStore(``)(*current) = nil, expected error")
	}
}
