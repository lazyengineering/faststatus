// Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/lazyengineering/faststatus/resource"
)

var testResources map[uint64]resource.Resource

func init() {
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

func TestCurrent(t *testing.T) {
	{ // zero value works
		c, err := Current()
		if err != nil {
			t.Fatalf("Current() = handler, %+v, expected nil error", err)
		}
		if c == nil {
			t.Fatalf("Current() = %+v, err, expected non-nil http.Handler", c)
		}
		if _, ok := c.(*current); !ok {
			t.Fatalf("Current() = %+v, err, expected *content as http.Handler", c)
		}
	}
	{ // noop options work
		c, err := Current(noopCurrentOption)
		if err != nil {
			t.Fatalf("Current(noopCurrentOption) = handler, %+v, expected nil error", err)
		}
		if c == nil {
			t.Fatalf("Current(noopCurrentOption) = %+v, err, expected non-nil http.Handler", c)
		}
		if _, ok := c.(*current); !ok {
			t.Fatalf("Current(noopCurrentOption) = %+v, err, expected *content as http.Handler", c)
		}
	}
	{ // option errors propagate
		c, err := Current(errCurrentOption)
		if err == nil {
			t.Fatalf("Current(errCurrentOption) = handler, %+v, expected non-nil error", err)
		}
		if c != nil {
			t.Fatalf("Current(errCurrentOption) = %+v, err, expected nil http.Handler", c)
		}
	}
	{ // options modify current
		s, err := Current(mockNoopStoreOption(1))
		if err != nil {
			t.Fatalf("Current(mockNoopStoreOption(1)) = handler, %+v, expected nil error", err)
		}
		if s == nil {
			t.Fatalf("Current(mockNoopStoreOption(1)) = %+v, err, expected non-nil http.Handler", s)
		}
		c, ok := s.(*current)
		if !ok {
			t.Fatalf("Current(mockNoopStoreOption(1)) = %+v, err, expected *content as http.Handler", s)
		}
		if c.store != mockNoopStore(1) {
			t.Fatalf("Current(mockNoopStoreOption(1)).store = %+v, expected %+v", c.store, mockNoopStore(1))
		}
	}
	{ // options overwrite earlier options
		s, err := Current(mockNoopStoreOption(1), mockNoopStoreOption(2))
		if err != nil {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(2)) = handler, %+v, expected nil error", err)
		}
		if s == nil {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(2)) = %+v, err, expected non-nil http.Handler", s)
		}
		c, ok := s.(*current)
		if !ok {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(2)) = %+v, err, expected *content as http.Handler", s)
		}
		if c.store != mockNoopStore(2) {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(2)).store = %+v, expected %+v", c.store, mockNoopStore(2))
		}
	}
	{ // idempotent options work
		s, err := Current(mockNoopStoreOption(1), mockNoopStoreOption(1))
		if err != nil {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(1)) = handler, %+v, expected nil error", err)
		}
		if s == nil {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(1)) = %+v, err, expected non-nil http.Handler", s)
		}
		c, ok := s.(*current)
		if !ok {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(1)) = %+v, err, expected *content as http.Handler", s)
		}
		if c.store != mockNoopStore(1) {
			t.Fatalf("Current(mockNoopStoreOption(1),mockNoopStoreOptions(1)).store = %+v, expected %+v", c.store, mockNoopStore(1))
		}
	}
}

func ExampleCurrent() {
	server, err := Current()
	if err != nil {
		log.Fatalf("unable to start Current service: %+v", err)
	}
	http.Handle("/resources/", http.StripPrefix("/resources/", server))
}

func TestCurrentServeHTTP_GET(t *testing.T) {
	c := &http.Client{}

	// test table
	type expectation struct {
		statusCode int
		body       string
	}
	type input struct {
		path   string
		accept []string
		store  store
	}
	type test struct {
		in      input
		expects expectation
	}
	jsonBody := func(rs ...resource.Resource) string {
		b, err := json.Marshal(rs)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	tests := []test{
		test{ // "bad request"
			input{
				path:   "/ghi",
				accept: append([]string{}, "*/*"),
				store:  &mockFailureStore{t: t},
			},
			expectation{
				statusCode: http.StatusBadRequest,
				body:       "Bad Request\n",
			},
		},
		test{ // "/" "*/*"
			input{
				path:   "/",
				accept: append([]string{}, "*/*"),
				store: &mockGetStore{
					ids:       []uint64{},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "Resource Not Found\n",
			},
		},
		test{ // "/1" "*/*"
			input{
				path:   "/1",
				accept: append([]string{}, "*/*"),
				store: &mockGetStore{
					ids:       []uint64{1},
					resources: []resource.Resource{testResources[1]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       fmt.Sprintln(testResources[1].String()),
			},
		},
		test{ // "/1/2/" "*/*"
			input{
				path:   "/1/2/",
				accept: append([]string{}, "*/*"),
				store: &mockGetStore{
					ids:       []uint64{1, 2},
					resources: []resource.Resource{testResources[1], testResources[2]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       fmt.Sprintf("%v\n%v\n", testResources[1].String(), testResources[2].String()),
			},
		},
		test{ // "/e" "*/*"
			input{
				path:   "/e",
				accept: append([]string{}, "*/*"),
				store: &mockGetStore{
					ids:       []uint64{14},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "Resource Not Found\n",
			},
		},
		test{ // "/1/2/a" "*/*"
			input{
				path:   "/1/2/a",
				accept: append([]string{}, "*/*"),
				store: &mockGetStore{
					ids:       []uint64{1, 2, 10},
					resources: []resource.Resource{testResources[1], testResources[2], testResources[10]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       fmt.Sprintf("%v\n%v\n%v\n", testResources[1].String(), testResources[2].String(), testResources[10].String()),
			},
		},
		test{ // "/" "text/plain"
			input{
				path:   "/",
				accept: append([]string{}, "text/plain"),
				store: &mockGetStore{
					ids:       []uint64{},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "Resource Not Found\n",
			},
		},
		test{ // "/1" "text/plain"
			input{
				path:   "/1",
				accept: append([]string{}, "text/plain"),
				store: &mockGetStore{
					ids:       []uint64{1},
					resources: []resource.Resource{testResources[1]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       fmt.Sprintln(testResources[1].String()),
			},
		},
		test{ // "/1/2/" "text/plain"
			input{
				path:   "/1/2/",
				accept: append([]string{}, "text/plain"),
				store: &mockGetStore{
					ids:       []uint64{1, 2},
					resources: []resource.Resource{testResources[1], testResources[2]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       fmt.Sprintf("%v\n%v\n", testResources[1].String(), testResources[2].String()),
			},
		},
		test{ // "/e" "text/plain"
			input{
				path:   "/e",
				accept: append([]string{}, "text/plain"),
				store: &mockGetStore{
					ids:       []uint64{14},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "Resource Not Found\n",
			},
		},
		test{ // "/1/2/a" "text/plain"
			input{
				path:   "/1/2/a",
				accept: append([]string{}, "text/plain"),
				store: &mockGetStore{
					ids:       []uint64{1, 2, 10},
					resources: []resource.Resource{testResources[1], testResources[2], testResources[10]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       fmt.Sprintf("%v\n%v\n%v\n", testResources[1].String(), testResources[2].String(), testResources[10].String()),
			},
		},
		test{ // "/" "application/json"
			input{
				path:   "/",
				accept: append([]string{}, "application/json"),
				store: &mockGetStore{
					ids:       []uint64{},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "[]\n",
			},
		},
		test{ // "/1" "application/json"
			input{
				path:   "/1",
				accept: append([]string{}, "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1},
					resources: []resource.Resource{testResources[1]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1]) + "\n",
			},
		},
		test{ // "/1/2/" "application/json"
			input{
				path:   "/1/2/",
				accept: append([]string{}, "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1, 2},
					resources: []resource.Resource{testResources[1], testResources[2]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
		test{ // "/e" "application/json"
			input{
				path:   "/e",
				accept: append([]string{}, "application/json"),
				store: &mockGetStore{
					ids:       []uint64{14},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "[]" + "\n",
			},
		},
		test{ // "/1/2/a" "application/json"
			input{
				path:   "/1/2/a",
				accept: append([]string{}, "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1, 2, 10},
					resources: []resource.Resource{testResources[1], testResources[2], testResources[10]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1], testResources[2], testResources[10]) + "\n",
			},
		},
		test{ // "/" "text/html,application/json"
			input{
				path:   "/",
				accept: append([]string{}, "text/html", "application/json"),
				store: &mockGetStore{
					ids:       []uint64{},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "[]\n",
			},
		},
		test{ // "/1" "text/html,application/json"
			input{
				path:   "/1",
				accept: append([]string{}, "text/html", "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1},
					resources: []resource.Resource{testResources[1]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1]) + "\n",
			},
		},
		test{ // "/1/2/" "text/html,application/json"
			input{
				path:   "/1/2/",
				accept: append([]string{}, "text/html", "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1, 2},
					resources: []resource.Resource{testResources[1], testResources[2]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
		test{ // "/e" "text/html,application/json"
			input{
				path:   "/e",
				accept: append([]string{}, "text/html", "application/json"),
				store: &mockGetStore{
					ids:       []uint64{14},
					resources: []resource.Resource{},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusNotFound,
				body:       "[]" + "\n",
			},
		},
		test{ // "/1/2/a" "text/html,application/json"
			input{
				path:   "/1/2/a",
				accept: append([]string{}, "text/html", "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1, 2, 10},
					resources: []resource.Resource{testResources[1], testResources[2], testResources[10]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1], testResources[2], testResources[10]) + "\n",
			},
		},
		test{ // "/1/2/e" "text/html,application/json"
			input{
				path:   "/1/2/e",
				accept: append([]string{}, "text/html", "application/json"),
				store: &mockGetStore{
					ids:       []uint64{1, 2, 14},
					resources: []resource.Resource{testResources[1], testResources[2]},
					err:       nil,
					t:         t,
				},
			},
			expectation{
				statusCode: http.StatusOK,
				body:       jsonBody(testResources[1], testResources[2]) + "\n",
			},
		},
	}
	for _, tst := range tests {
		s := httptest.NewServer(&current{tst.in.store})

		var b bytes.Buffer
		rq, err := http.NewRequest(http.MethodGet, s.URL+tst.in.path, &b)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range tst.in.accept {
			rq.Header.Add("Accept", a)
		}
		if r, err := c.Do(rq); err != nil {
			t.Error(err)
		} else {
			if r.StatusCode != tst.expects.statusCode {
				t.Errorf("GET %s (%s) = %000d, expected %000d\n", tst.in.path, tst.in.accept, r.StatusCode, tst.expects.statusCode)
			}
			body, er := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if er != nil {
				t.Fatal(er)
			}
			if string(body) != tst.expects.body {
				t.Errorf("GET %s (%s) = %q, expected %q\n", tst.in.path, tst.in.accept, string(body), tst.expects.body)
			}
		}
	}
}

func mustParse(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return t
}

func noopCurrentOption(c *current) error {
	return nil
}

func errCurrentOption(c *current) error {
	return fmt.Errorf("BANG!")
}

// make it nice and comparable...
type mockNoopStore int

func (s mockNoopStore) save(r resource.Resource) error {
	return nil
}

func (s mockNoopStore) get(ids ...uint64) ([]resource.Resource, error) {
	return nil, nil
}

func mockNoopStoreOption(n int) func(*current) error {
	return func(c *current) error {
		c.store = mockNoopStore(n)
		return nil
	}
}

type mockErrorStore string

func (s mockErrorStore) Error() string {
	return string(s)
}

func (s mockErrorStore) String() string {
	return string(s)
}

func (s mockErrorStore) save(r resource.Resource) error {
	return s
}

func (s mockErrorStore) get(ids ...uint64) ([]resource.Resource, error) {
	return nil, s
}

// attaches a mockErrorStore, doesn't return an error...
func mockErrorStoreOption(e string) func(*current) error {
	return func(c *current) error {
		c.store = mockErrorStore(e)
		return nil
	}
}

type mockGetStore struct {
	ids       []uint64
	resources []resource.Resource
	err       error
	t         *testing.T
}

func (s *mockGetStore) save(r resource.Resource) error {
	return nil
}

func (s *mockGetStore) get(ids ...uint64) ([]resource.Resource, error) {
	if !(reflect.DeepEqual(ids, s.ids) || (len(ids) == len(s.ids) && len(ids) == 0)) {
		s.t.Errorf("store.get(%+v), expected store.get(%+v)", ids, s.ids)
	}
	return s.resources, s.err
}

type mockSaveStore struct {
	resource resource.Resource
	err      error
	t        *testing.T
}

func (s *mockSaveStore) save(r resource.Resource) error {
	if !reflect.DeepEqual(r, s.resource) {
		s.t.Errorf("store.save(%+v), expected store.save(%+v)", r, s.resource)
	}
	return s.err
}

func (s *mockSaveStore) get(ids ...uint64) ([]resource.Resource, error) {
	return nil, nil
}

// will always fail a test when methods are called
type mockFailureStore struct {
	t *testing.T
}

func (s *mockFailureStore) save(r resource.Resource) error {
	s.t.Fatal("unexpected call to store.save()")
	return nil
}

func (s *mockFailureStore) get(ids ...uint64) ([]resource.Resource, error) {
	s.t.Fatal("unexpected call to store.get()")
	return nil, nil
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
	if err := populateDb(); err != nil {
		log.Fatal(err)
	}
}
