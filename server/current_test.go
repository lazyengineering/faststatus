// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package server

import (
	"bytes"
	"encoding/json"
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
func TestCurrent(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		log.Fatal(err)
	}
	fnm := tmpfile.Name()
	err = tmpfile.Close()
	if err != nil {
		log.Fatal(err)
	}
	c, er := Current(fnm)
	if er != nil {
		t.Fatalf("Unexpected error: %v", er)
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
	setupTest()
	defer teardownDb()

	s := httptest.NewServer(mustHandler(Current(db_file)))
	c := &http.Client{}

	// test table
	type expectation struct {
		StatusCode int
		Body       string
	}
	type input struct {
		Path   string
		Accept []string
	}
	type test struct {
		I input
		E expectation
	}
	jsonBody := func(rs ...resource.Resource) string {
		b, err := json.Marshal(rs)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	tests := []test{
		test{ // "/" "*/*"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1" "*/*"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintln(testResources[1].String()),
			},
		},
		test{ // "/1/2/" "*/*"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n", testResources[1].String(), testResources[2].String()),
			},
		},
		test{ // "/e" "*/*"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1/2/a" "*/*"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n%v\n", testResources[1].String(), testResources[2].String(), testResources[10].String()),
			},
		},
		test{ // "/" "text/plain"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1" "text/plain"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintln(testResources[1].String()),
			},
		},
		test{ // "/1/2/" "text/plain"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n", testResources[1].String(), testResources[2].String()),
			},
		},
		test{ // "/e" "text/plain"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1/2/a" "text/plain"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n%v\n", testResources[1].String(), testResources[2].String(), testResources[10].String()),
			},
		},
		test{ // "/" "application/json"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]\n",
			},
		},
		test{ // "/1" "application/json"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1]) + "\n",
			},
		},
		test{ // "/1/2/" "application/json"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
		test{ // "/e" "application/json"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]" + "\n",
			},
		},
		test{ // "/1/2/a" "application/json"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2], testResources[10]) + "\n",
			},
		},
		test{ // "/" "text/html,application/json"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]\n",
			},
		},
		test{ // "/1" "text/html,application/json"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1]) + "\n",
			},
		},
		test{ // "/1/2/" "text/html,application/json"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
		test{ // "/e" "text/html,application/json"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]" + "\n",
			},
		},
		test{ // "/1/2/a" "text/html,application/json"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2], testResources[10]) + "\n",
			},
		},
	}
	for _, tst := range tests {
		var b bytes.Buffer
		rq, err := http.NewRequest(http.MethodGet, s.URL+tst.I.Path, &b)
		if err != nil {
			t.Error(err)
		}
		for _, a := range tst.I.Accept {
			rq.Header.Add("Accept", a)
		}
		if r, err := c.Do(rq); err != nil {
			t.Error(err)
		} else {
			if r.StatusCode != tst.E.StatusCode {
				t.Errorf("Expected: %v\tActual: %v\n", tst.E.StatusCode, r.StatusCode)
			}
			body, er := ioutil.ReadAll(r.Body)
			if er != nil {
				t.Error(er)
			}
			if string(body) != tst.E.Body {
				t.Errorf("%v,%v\t Expected: %v\tActual: %v\n", tst.I.Path, tst.I.Accept, tst.E.Body, string(body))
			}
		}
	}
}

//TODO: Enable test again when ready to start failing...
func testCurrentServeHTTP_PUT(t *testing.T) {
	setupTest()
	defer teardownDb()

	s := httptest.NewServer(mustHandler(Current(db_file)))
	c := &http.Client{}

	t.Fatal("Not Implemented")

	// test table
	type expectation struct {
		StatusCode int
		Body       string
	}
	type input struct {
		Path   string
		Accept []string
		Body   string
	}
	type test struct {
		I input
		E expectation
	}
	jsonBody := func(rs ...resource.Resource) string {
		b, err := json.Marshal(rs)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	tests := []test{
		test{ // "/" "*/*"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1" "*/*"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintln(testResources[1].String()),
			},
		},
		test{ // "/1/2/" "*/*"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n", testResources[1].String(), testResources[2].String()),
			},
		},
		test{ // "/e" "*/*"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1/2/a" "*/*"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "*/*"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n%v\n", testResources[1].String(), testResources[2].String(), testResources[10].String()),
			},
		},
		test{ // "/" "text/plain"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1" "text/plain"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintln(testResources[1].String()),
			},
		},
		test{ // "/1/2/" "text/plain"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n", testResources[1].String(), testResources[2].String()),
			},
		},
		test{ // "/e" "text/plain"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "Resource Not Found\n",
			},
		},
		test{ // "/1/2/a" "text/plain"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "text/plain"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       fmt.Sprintf("%v\n%v\n%v\n", testResources[1].String(), testResources[2].String(), testResources[10].String()),
			},
		},
		test{ // "/" "application/json"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]\n",
			},
		},
		test{ // "/1" "application/json"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1]) + "\n",
			},
		},
		test{ // "/1/2/" "application/json"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
		test{ // "/e" "application/json"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]" + "\n",
			},
		},
		test{ // "/1/2/a" "application/json"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2], testResources[10]) + "\n",
			},
		},
		test{ // "/" "text/html,application/json"
			I: input{
				Path:   "/",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]\n",
			},
		},
		test{ // "/1" "text/html,application/json"
			I: input{
				Path:   "/1",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1]) + "\n",
			},
		},
		test{ // "/1/2/" "text/html,application/json"
			I: input{
				Path:   "/1/2/",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
		test{ // "/e" "text/html,application/json"
			I: input{
				Path:   "/e",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusNotFound,
				Body:       "[]" + "\n",
			},
		},
		test{ // "/1/2/a" "text/html,application/json"
			I: input{
				Path:   "/1/2/a",
				Accept: append([]string{}, "text/html", "application/json"),
			},
			E: expectation{
				StatusCode: http.StatusOK,
				Body:       jsonBody(testResources[1], testResources[2], testResources[10]) + "\n",
			},
		},
	}
	for _, tst := range tests {
		var b bytes.Buffer
		rq, err := http.NewRequest(http.MethodGet, s.URL+tst.I.Path, &b)
		if err != nil {
			t.Error(err)
		}
		for _, a := range tst.I.Accept {
			rq.Header.Add("Accept", a)
		}
		if r, err := c.Do(rq); err != nil {
			t.Error(err)
		} else {
			if r.StatusCode != tst.E.StatusCode {
				t.Errorf("Expected: %v\tActual: %v\n", tst.E.StatusCode, r.StatusCode)
			}
			body, er := ioutil.ReadAll(r.Body)
			if er != nil {
				t.Error(er)
			}
			if string(body) != tst.E.Body {
				t.Errorf("%v,%v\t Expected: %v\tActual: %v\n", tst.I.Path, tst.I.Accept, tst.E.Body, string(body))
			}
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

func setupTest() {
	setupDb()
	setupResources()
	if err := populateDb(); err != nil {
		log.Fatal(err)
	}
}
