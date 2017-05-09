// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest

import (
	"net/http"

	"github.com/lazyengineering/faststatus"
)

// Server is a restful http server for Resources.
type Server struct {
}

// ServerOpt is used to configure a Server
type ServerOpt func(*Server) error

// New provides a restful endpoint for managing faststatus Resources.
func New(opts ...ServerOpt) (*Server, error) {
	s := &Server{}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/new":
		s.handleNew(w, r)
	case "/":
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		s.handleResource(w, r)
	}
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) handleResource(w http.ResponseWriter, r *http.Request) {
	var id faststatus.ID
	if err := (&id).UnmarshalText([]byte(r.URL.Path[1:])); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	case http.MethodPut:
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}
