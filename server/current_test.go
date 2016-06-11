// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package server

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/lazyengineering/faststatus/resource"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

// New Server for current status
func TestNewCurrent(t *testing.T) {
	c, err := NewCurrent("tmp-db.db")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("Expected server")
	}
}

func mustHandler(h http.Handler, e error) http.Handler {
	if e != nil {
		panic(e)
	}
	return h
}

func TestCurrentServeHTTP_GET(t *testing.T) {
	s := httptest.NewServer(mustHandler(NewCurrent(db_file)))
	c := &http.Client{}
	if r, err := c.Get(s.URL + "/"); err != nil {
		t.Error(err)
	} else {
		if r.StatusCode != http.StatusNotFound {
			t.Error("Expected: StatusNotFound\tActual: ", r.StatusCode)
		}
	}
	if r, err := c.Get(s.URL + "/1"); err != nil {
		t.Error(err)
	} else {
		if r.StatusCode != http.StatusOK {
			t.Error("Expected: StatusOK\tActual: ", r.StatusCode)
			t.Fail()
		}
		body, er := ioutil.ReadAll(r.Body)
		if er != nil {
			t.Error(er)
		}
		expected := fmt.Sprintln(testResources[1].String())
		if string(body) != expected {
			t.Errorf("Expected: %q\tActual: %q", expected, body)
		}
	}
	if r, err := c.Get(s.URL + "/1/2/"); err != nil {
		t.Error(err)
	} else {
		if r.StatusCode != http.StatusOK {
			t.Error("Expected: StatusOK\tActual: ", r.StatusCode)
		}
		body, er := ioutil.ReadAll(r.Body)
		if er != nil {
			t.Error(er)
		}
		expected := fmt.Sprintln(testResources[1].String())
		expected += fmt.Sprintln(testResources[2].String())
		if string(body) != expected {
			t.Errorf("Expected: %q\tActual: %q", expected, body)
		}
	}
	if r, err := c.Get(s.URL + "/e"); err != nil {
		t.Error(err)
	} else {
		if r.StatusCode != http.StatusNotFound {
			t.Error("Expected: StatusNotFound\tActual: ", r.StatusCode)
		}
	}
	if r, err := c.Get(s.URL + "/1/2/a"); err != nil {
		t.Error(err)
	} else {
		if r.StatusCode != http.StatusOK {
			t.Error("Expected: StatusOK\tActual: ", r.StatusCode)
		}
		body, er := ioutil.ReadAll(r.Body)
		if er != nil {
			t.Error(er)
		}
		expected := fmt.Sprintln(testResources[1].String())
		expected += fmt.Sprintln(testResources[2].String())
		expected += fmt.Sprintln(testResources[10].String())
		if string(body) != expected {
			t.Errorf("Expected: %q\tActual: %q", expected, body)
		}
	}
}

var db_file string

func setupDb() {
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		log.Fatal(err)
	}
	db_file = tmpfile.Name()
	err = tmpfile.Close()
	if err != nil {
		log.Fatal(err)
	}
}

var testResources map[uint64]resource.Resource

func mustParse(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

func setupResources() {
	testResources = map[uint64]resource.Resource{
		1: resource.Resource{
			Id:           1,
			FriendlyName: "First",
			Status:       resource.Free,
			Since:        mustParse(time.Parse(time.RFC3339, "2016-06-10T16:42:00Z")),
		},
		2: resource.Resource{
			Id:           2,
			FriendlyName: "Second",
			Status:       resource.Busy,
			Since:        mustParse(time.Parse(time.RFC3339, "2016-06-10T16:52:00Z")),
		},
		3: resource.Resource{
			Id:           3,
			FriendlyName: "Third",
			Status:       resource.Occupied,
			Since:        mustParse(time.Parse(time.RFC3339, "2016-06-10T17:02:00Z")),
		},
		10: resource.Resource{
			Id:           10,
			FriendlyName: "Tenth",
			Status:       resource.Busy,
			Since:        mustParse(time.Parse(time.RFC3339, "2016-06-10T17:12:00Z")),
		},
	}
}

func populateDb() error {
	db, err := bolt.Open(db_file, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte("resources")) != nil {
			tx.DeleteBucket([]byte("resources"))
		}

		b, err := tx.CreateBucket([]byte("resources"))
		if err != nil {
			return err
		}

		for _, r := range testResources {
			j, err := r.MarshalJSON()
			if err != nil {
				return err
			}
			b.Put([]byte(strconv.FormatUint(r.Id, 16)), j)
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func teardownDb() {
	os.Remove(db_file)
}

func TestMain(m *testing.M) {
	setupDb()
	defer teardownDb()

	setupResources()

	if err := populateDb(); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
