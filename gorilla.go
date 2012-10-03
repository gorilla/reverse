// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// GorillaHost ----------------------------------------------------------------

func NewGorillaHost(pattern string) (*GorillaHost, error) {
	pattern, err := gorillaPattern(pattern, true, false, false)
	if err != nil {
		return nil, err
	}
	r, err := CompileRegexp(pattern)
	if err != nil {
		return nil, err
	}
	return &GorillaHost{*r}, nil
}

// GorillaHost matches a URL host using Gorilla's special syntax for named
// groups: `{name:regexp}`.
type GorillaHost struct {
	Regexp
}

func (m *GorillaHost) Match(r *http.Request) bool {
	return m.MatchString(getHost(r))
}

// Extract returns positional and named variables extracted from the URL host.
func (m *GorillaHost) Extract(result *Result, r *http.Request) {
	result.Values = mergeValues(result.Values, m.Values(getHost(r)))
}

// Build builds the URL host using the given positional and named variables,
// and writes it to the given URL.
func (m *GorillaHost) Build(u *url.URL, values url.Values) error {
	host, err := m.RevertValid(values)
	if err == nil {
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		u.Host = host
	}
	return err
}

// GorillaPath ----------------------------------------------------------------

func NewGorillaPath(pattern string, strictSlash bool) (*GorillaPath, error) {
	regexpPattern, err := gorillaPattern(pattern, false, false, strictSlash)
	if err != nil {
		return nil, err
	}
	r, err := CompileRegexp(regexpPattern)
	if err != nil {
		return nil, err
	}
	return &GorillaPath{*r, pattern, strictSlash}, nil
}

// GorillaPath matches a URL path using Gorilla's special syntax for named
// groups: `{name:regexp}`.
type GorillaPath struct {
	Regexp
	pattern     string
	strictSlash bool
}

func (m *GorillaPath) Match(r *http.Request) bool {
	return m.MatchString(r.URL.Path)
}

// Extract returns positional and named variables extracted from the URL path.
func (m *GorillaPath) Extract(result *Result, r *http.Request) {
	result.Values = mergeValues(result.Values, m.Values(r.URL.Path))
	if result.Handler == nil && m.strictSlash {
		result.Handler = redirectPath(m.pattern, r)
	}
}

// Build builds the URL path using the given positional and named variables,
// and writes it to the given URL.
func (m *GorillaPath) Build(u *url.URL, values url.Values) error {
	path, err := m.RevertValid(values)
	if err == nil {
		u.Path = path
	}
	return err
}

// GorillaPathPrefix ----------------------------------------------------------

func NewGorillaPathPrefix(pattern string) (*GorillaPathPrefix, error) {
	regexpPattern, err := gorillaPattern(pattern, false, true, false)
	if err != nil {
		return nil, err
	}
	r, err := CompileRegexp(regexpPattern)
	if err != nil {
		return nil, err
	}
	return &GorillaPathPrefix{*r}, nil
}

// GorillaPathPrefix matches a URL path prefix using Gorilla's special syntax
// for named groups: `{name:regexp}`.
type GorillaPathPrefix struct {
	Regexp
}

func (m *GorillaPathPrefix) Match(r *http.Request) bool {
	return m.MatchString(r.URL.Path)
}

// Extract returns positional and named variables extracted from the URL path.
func (m *GorillaPathPrefix) Extract(result *Result, r *http.Request) {
	result.Values = mergeValues(result.Values, m.Values(r.URL.Path))
}

// Build builds the URL path using the given positional and named variables,
// and writes it to the given URL.
func (m *GorillaPathPrefix) Build(u *url.URL, values url.Values) error {
	path, err := m.RevertValid(values)
	if err == nil {
		u.Path = path
	}
	return err
}

// Helpers --------------------------------------------------------------------

// gorillaPattern transforms a gorilla pattern into a regexp pattern.
func gorillaPattern(tpl string, matchHost, prefixMatch, strictSlash bool) (string, error) {
	// Check if it is well-formed.
	idxs, err := braceIndices(tpl)
	if err != nil {
		return "", err
	}
	// Now let's parse it.
	defaultPattern := "[^/]+"
	if matchHost {
		defaultPattern = "[^.]+"
		prefixMatch, strictSlash = false, false
	} else {
		if prefixMatch {
			strictSlash = false
		}
		if strictSlash && strings.HasSuffix(tpl, "/") {
			tpl = tpl[:len(tpl)-1]
		}
	}
	pattern := bytes.NewBufferString("^")
	var end int
	for i := 0; i < len(idxs); i += 2 {
		// Set all values we are interested in.
		raw := tpl[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(tpl[idxs[i]+1:end-1], ":", 2)
		name := parts[0]
		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}
		// Name or pattern can't be empty.
		if name == "" || patt == "" {
			return "", fmt.Errorf("missing name or pattern in %q",
				tpl[idxs[i]:end])
		}
		// Build the regexp pattern.
		fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), name, patt)
	}
	// Add the remaining.
	raw := tpl[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	if strictSlash {
		pattern.WriteString("[/]?")
	}
	if !prefixMatch {
		pattern.WriteByte('$')
	}
	return pattern.String(), nil
}

// braceIndices returns the first level curly brace indices from a string.
// It returns an error in case of unbalanced braces.
func braceIndices(s string) ([]int, error) {
	var level, idx int
	idxs := make([]int, 0)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}
