package flow

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

var allMethods = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}

type contextKey string

func Param(ctx context.Context, param string) string {
	return ctx.Value(contextKey(param)).(string)
}

type Mux struct {
	NotFound         http.Handler
	MethodNotAllowed http.Handler
	Options          http.Handler
	routes           *[]route
	middlewares      []func(http.Handler) http.Handler
}

func New() *Mux {
	return &Mux{
		NotFound: http.NotFoundHandler(),
		MethodNotAllowed: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}),
		Options: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
		routes: &[]route{},
	}
}

func (m *Mux) Handle(pattern string, handler http.Handler, methods ...string) {
	if contains(methods, http.MethodGet) && !contains(methods, http.MethodHead) {
		methods = append(methods, http.MethodHead)
	}

	if len(methods) == 0 {
		methods = allMethods
	}

	for _, method := range methods {
		route := route{
			method:   strings.ToUpper(method),
			segments: strings.Split(pattern, "/"),
			wildcard: strings.HasSuffix(pattern, "/..."),
			handler:  m.wrap(handler),
		}

		*m.routes = append(*m.routes, route)
	}
}

func (m *Mux) HandleFunc(pattern string, fn http.HandlerFunc, methods ...string) {
	m.Handle(pattern, fn, methods...)
}

func (m *Mux) Use(mw func(http.Handler) http.Handler) {
	m.middlewares = append(m.middlewares, mw)
}

func (m *Mux) Group(fn func(*Mux)) {
	mm := *m
	fn(&mm)
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlSegments := strings.Split(r.URL.Path, "/")
	allowedMethods := []string{}

	for _, route := range *m.routes {
		ctx, ok := route.match(r.Context(), urlSegments)
		if ok {
			if r.Method == route.method {
				route.handler.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if !contains(allowedMethods, route.method) {
				allowedMethods = append(allowedMethods, route.method)
			}
		}
	}

	if len(allowedMethods) > 0 {
		w.Header().Set("Allow", strings.Join(append(allowedMethods, http.MethodOptions), ", "))
		if r.Method == http.MethodOptions {
			m.wrap(m.Options).ServeHTTP(w, r)
		} else {
			m.wrap(m.MethodNotAllowed).ServeHTTP(w, r)
		}
		return
	}

	m.wrap(m.NotFound).ServeHTTP(w, r)
}

func (m *Mux) wrap(handler http.Handler) http.Handler {
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}

	return handler
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type route struct {
	method   string
	segments []string
	wildcard bool
	handler  http.Handler
}

func (r *route) match(ctx context.Context, urlSegments []string) (context.Context, bool) {
	if !r.wildcard && len(urlSegments) != len(r.segments) {
		return ctx, false
	}

	for i, routeSegment := range r.segments {
		if i > len(urlSegments)-1 {
			return ctx, false
		}

		if routeSegment == "..." {
			return ctx, true
		}

		if strings.HasPrefix(routeSegment, ":") {
			pipe := strings.Index(routeSegment, "|")
			if pipe == -1 {
				ctx = context.WithValue(ctx, contextKey(routeSegment), urlSegments[i])
				continue
			}

			rx := regexp.MustCompile(routeSegment[pipe+1:])
			if rx.MatchString(urlSegments[i]) {
				ctx = context.WithValue(ctx, contextKey(routeSegment[:pipe]), urlSegments[i])
				continue
			}

			return ctx, false
		}

		if urlSegments[i] != routeSegment {
			return ctx, false
		}
	}

	return ctx, true
}
