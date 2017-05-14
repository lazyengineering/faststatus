// Copyright 2017 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package rest

import (
	"fmt"
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
	err := s.serveHTTP(w, r)
	if err != nil {
		http.Error(w, http.StatusText(errorCode(err)), errorCode(err))
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	switch r.URL.Path {
	case "/":
		return &restError{code: http.StatusNotFound}
	case "/new":
		return s.handleNew(w, r)
	default:
		return s.handleResource(w, r)
	}
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	default:
		return &restError{code: http.StatusMethodNotAllowed}
	}
	resource := faststatus.NewResource()
	txt, err := resource.MarshalText()
	if err != nil {
		return fmt.Errorf("marshaling to text: %+v", err)
	}
	w.Write(txt)
	return nil
}

func (s *Server) handleResource(w http.ResponseWriter, r *http.Request) error {
	var id faststatus.ID
	if err := (&id).UnmarshalText([]byte(r.URL.Path[1:])); err != nil {
		return &restError{
			err:  fmt.Errorf("unmarshalling id from path: %+v", err),
			code: http.StatusNotFound,
		}
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return s.getResource(id).serveHTTP(w, r)
	case http.MethodPut:
		return s.putResource(id).serveHTTP(w, r)
	default:
		return &restError{code: http.StatusMethodNotAllowed}
	}
}

func (s *Server) putResource(id faststatus.ID) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("reading from request body: %+v", err)
		}
		resource := new(faststatus.Resource)
		if err := resource.UnmarshalText(b); err != nil {
			return &restError{
				err:  fmt.Errorf("unmarshaling resource from request: %+v", err),
				code: http.StatusBadRequest,
			}
		}
		if resource.Since.IsZero() {
			return &restError{
				err:  fmt.Errorf("zero-value Since"),
				code: http.StatusBadRequest,
			}
		}
		if id != resource.ID {
			return &restError{
				err:  fmt.Errorf("resource ID %q does not match path ID %q", resource.ID, id),
				code: http.StatusBadRequest,
			}
		}
		if err := s.store.Save(*resource); faststatus.ConflictError(err) {
			return &restError{
				err:  err,
				code: http.StatusConflict,
			}
		} else if err != nil {
			return fmt.Errorf("saving resource to store: %+v", err)
		}
		rb, err := resource.MarshalText()
		if err != nil {
			return fmt.Errorf("marshaling resource for response: %+v", err)
		}
		w.Write(rb)
		return nil
	}
}

func (s *Server) getResource(id faststatus.ID) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		resource, err := s.store.Get(id)
		if err != nil {
			return fmt.Errorf("getting resource from store: %+v", err)
		}
		if resource.Equal(faststatus.Resource{}) {
			return &restError{code: http.StatusNotFound}
		}
		rb, err := resource.MarshalText()
		if err != nil {
			return fmt.Errorf("marshaling resource for response: %+v", err)
		}
		w.Write(rb)
		return nil
	}
}

type handler interface {
	serveHTTP(http.ResponseWriter, *http.Request) error
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func (fn handlerFunc) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	return fn(w, r)
}
