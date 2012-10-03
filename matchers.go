// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"net/http"
	"net/url"
	"strings"
)

// Result stores the results from a match.
type Result struct {
	Handler http.Handler
	Values  url.Values
}

// Matcher matches a request.
type Matcher interface {
	Match(*http.Request) bool
}

// Extractor extracts variables from a request.
type Extractor interface {
	Extract(*Result, *http.Request)
}

// Builder builds a URL based on positional and/or named variables.
type Builder interface {
	Build(*url.URL, url.Values) error
}

// Func -----------------------------------------------------------------------

// Func is a function signature for custom matchers.
type Func func(*http.Request) bool

func (m Func) Match(r *http.Request) bool {
	return m(r)
}

// Header ---------------------------------------------------------------------

// NewHeader returns a header matcher, converting keys to the canonical form.
func NewHeader(m map[string]string) Header {
	for k, v := range m {
		delete(m, k)
		m[http.CanonicalHeaderKey(k)] = v
	}
	return Header(m)
}

// Header matches request headers. All values, if non-empty, must match.
// Empty values only check if the header is present.
type Header map[string]string

func (m Header) Match(r *http.Request) bool {
	src := r.Header
loop:
	for k, v := range m {
		if values, ok := src[k]; !ok {
			return false
		} else if v != "" {
			for _, value := range values {
				if v == value {
					continue loop
				}
			}
			return false
		}
	}
	return true
}

// Host -----------------------------------------------------------------------

// NewHost returns a static URL host matcher.
func NewHost(host string) Host {
	return Host(host)
}

// Host matches a static URL host.
type Host string

func (m Host) Match(r *http.Request) bool {
	return getHost(r) == string(m)
}

// Method ---------------------------------------------------------------------

// NewMethod retuns a request method matcher, converting values to upper-case.
func NewMethod(m []string) Method {
	for k, v := range m {
		m[k] = strings.ToUpper(v)
	}
	return Method(m)
}

// Method matches the request method. One of the values must match.
type Method []string

func (m Method) Match(r *http.Request) bool {
	for _, v := range m {
		if v == r.Method {
			return true
		}
	}
	return false
}

// None -----------------------------------------------------------------------

// NewNone returns a matcher that never matches.
func NewNone() *None {
	return nil
}

// None never matches.
type None bool

func (m *None) Match(r *http.Request) bool {
	return false
}

// Path -----------------------------------------------------------------------

// NewPath returns a static URL path matcher.
func NewPath(path string) Path {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return Path(path)
}

// Path matches a static URL path.
type Path string

func (m Path) Match(r *http.Request) bool {
	return r.URL.Path == string(m)
}

// PathRedirect ---------------------------------------------------------------

// NewPathRedirect returns a static URL path matcher that redirects if the
// trailing slash differs.
func NewPathRedirect(path string) PathRedirect {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return PathRedirect(path)
}

// PathRedirect matches a static URL path and redirects to the trailing-slash
// or non-trailing-slash version if it differs.
type PathRedirect string

func (m PathRedirect) Match(r *http.Request) bool {
	return strings.TrimRight(r.URL.Path, "/") == strings.TrimRight(string(m), "/")
}

func (m PathRedirect) Extract(result *Result, r *http.Request) {
	if result.Handler == nil {
		result.Handler = redirectPath(string(m), r)
	}
}

// PathPrefix -----------------------------------------------------------------

// NewPathPrefix returns a static URL path prefix matcher.
func NewPathPrefix(prefix string) PathPrefix {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return PathPrefix(prefix)
}

// PathPrefix matches a static URL path prefix.
type PathPrefix string

func (m PathPrefix) Match(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, string(m))
}

// Query ----------------------------------------------------------------------

// NewQuery returns a URL query matcher.
func NewQuery(m map[string]string) Query {
	return Query(m)
}

// Query matches URL queries. All values, if non-empty, must match.
// Empty values only check if the query is present.
type Query map[string]string

func (m Query) Match(r *http.Request) bool {
	src := r.URL.Query()
loop:
	for k, v := range m {
		if values, ok := src[k]; !ok {
			return false
		} else if v != "" {
			for _, value := range values {
				if v == value {
					continue loop
				}
			}
			return false
		}
	}
	return true
}

// Scheme ---------------------------------------------------------------------

// NewScheme retuns a URL scheme matcher, converting values to lower-case.
func NewScheme(m []string) Scheme {
	for k, v := range m {
		m[k] = strings.ToLower(v)
	}
	return Scheme(m)
}

// Scheme matches the URL scheme. One of the values must match.
type Scheme []string

func (m Scheme) Match(r *http.Request) bool {
	for _, v := range m {
		if v == r.URL.Scheme {
			return true
		}
	}
	return false
}

// Helpers --------------------------------------------------------------------

// getHost tries its best to return the request host.
func getHost(r *http.Request) string {
	if r.URL.IsAbs() {
		host := r.Host
		// Slice off any port information.
		if i := strings.Index(host, ":"); i != -1 {
			host = host[:i]
		}
		return host
	}
	return r.URL.Host
}

// mergeValues returns the result of merging two url.Values.
func mergeValues(u1, u2 url.Values) url.Values {
	if u1 == nil {
		return u2
	}
	if u2 == nil {
		return u1
	}
	for k, v := range u2 {
		u1[k] = append(u1[k], v...)
	}
	return u1
}

// redirectPath returns a handler that redirects if the path trailing slash
// differs from the request URL path.
func redirectPath(path string, r *http.Request) http.Handler {
	t1 := strings.HasSuffix(path, "/")
	t2 := strings.HasSuffix(r.URL.Path, "/")
	if t1 != t2 {
		u, _ := url.Parse(r.URL.String())
		if t1 {
			u.Path += "/"
		} else {
			u.Path = u.Path[:len(u.Path)-1]
		}
		return http.RedirectHandler(u.String(), 301)
	}
	return nil
}
