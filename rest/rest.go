// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest

import (
	"net/http"

	"github.com/lazyengineering/faststatus"
)

type server struct{}

// New provides a restful endpoint for managing faststatus Resources
func New() http.Handler {
	return &server{}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/new":
		s.handleNew(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (s *server) handleNew(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
	resource := faststatus.NewResource()
	txt, err := resource.MarshalText()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Write(txt)
}
