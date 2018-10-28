package http

import (
	"net/http"
	"strings"
)

type router struct {
	routes []*route
}

type route struct {
	method      string
	path        string
	prefix      bool
	handler     appHandler
	matcherFunc func(*http.Request) bool
}

func NewRouter() *router {
	return &router{}
}

func (r *router) Route(method, path string, handler appHandler) *route {
	route := route{
		method:  method,
		path:    path,
		handler: handler,
	}
	r.routes = append(r.routes, &route)
	return &route
}

func (r *router) RoutePrefix(method, path string, handler appHandler) *route {
	route := r.Route(method, path, handler)
	route.prefix = true
	return route
}

func (r *router) Handler() http.Handler {
	return appHandler(func(w http.ResponseWriter, req *http.Request) *appError {
		for _, route := range r.routes {
			if route.match(req) {
				return route.handler(w, req)
			}
		}
		return NotFoundHandler(w, req)
	})
}

func (r *route) Header(header, value string) {
	r.MatcherFunc(func(req *http.Request) bool {
		return req.Header.Get(header) == value
	})
}

func (r *route) MatcherFunc(f func(*http.Request) bool) {
	r.matcherFunc = f
}

func (r *route) match(req *http.Request) bool {
	if req.Method != r.method {
		return false
	}
	if r.prefix {
		if !strings.HasPrefix(req.URL.Path, r.path) {
			return false
		}
	} else if r.path != req.URL.Path {
		return false
	}
	return r.matcherFunc == nil || r.matcherFunc(req)
}
