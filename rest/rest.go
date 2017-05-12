// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/lazyengineering/faststatus"
)

// Server is a restful http server for Resources.
type Server struct {
	store Store
}

// Store gets and saves Resources.
type Store interface {
	Save(faststatus.Resource) error
	Get(faststatus.ID) (faststatus.Resource, error)
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

// WithStore configures a Server to use the provided Store.
func WithStore(store Store) ServerOpt {
	return func(s *Server) error {
		s.store = store
		return nil
	}
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
		return
	}
	resource := faststatus.NewResource()
	txt, err := resource.MarshalText()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Write(txt)
}

func (s *Server) handleResource(w http.ResponseWriter, r *http.Request) {
	var id faststatus.ID
	if err := (&id).UnmarshalText([]byte(r.URL.Path[1:])); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		s.getResource(id).ServeHTTP(w, r)
	case http.MethodPut:
		s.putResource(id).ServeHTTP(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *Server) putResource(id faststatus.ID) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		resource := new(faststatus.Resource)
		if err := resource.UnmarshalText(b); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if resource.Since.IsZero() {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if id != resource.ID {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err := s.store.Save(*resource); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		rb, err := resource.MarshalText()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Write(rb)
	})
}

func (s *Server) getResource(id faststatus.ID) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resource, err := s.store.Get(id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if resource.Equal(faststatus.Resource{}) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		rb, err := resource.MarshalText()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Write(rb)
	})
}
