// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"net/http"
	"net/url"
)

// RegexpHost -----------------------------------------------------------------

// NewRegexpHost returns a regexp matcher for the given URL host pattern.
func NewRegexpHost(pattern string) (*RegexpHost, error) {
	r, err := CompileRegexp(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexpHost{*r}, nil
}

// RegexpHost matches the URL host against a regular expression.
// The outermost capturing groups are extracted and the host can be reverted.
type RegexpHost struct {
	Regexp
}

func (m *RegexpHost) Match(r *http.Request) bool {
	return m.MatchString(getHost(r))
}

// Extract returns positional and named variables extracted from the URL host.
func (m *RegexpHost) Extract(result *Result, r *http.Request) {
	result.Values = mergeValues(result.Values, m.Values(getHost(r)))
}

// Build builds the URL host using the given positional and named variables,
// and writes it to the given URL.
func (m *RegexpHost) Build(u *url.URL, values url.Values) error {
	host, err := m.RevertValid(values)
	if err == nil {
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		u.Host = host
	}
	return err
}

// RegexpPath -----------------------------------------------------------------

// NewRegexpPath returns a regexp matcher for the given URL path pattern.
func NewRegexpPath(pattern string) (*RegexpPath, error) {
	r, err := CompileRegexp(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexpPath{*r}, nil
}

// RegexpPath matches the URL path against a regular expression.
// The outermost capturing groups are extracted and the path can be reverted.
type RegexpPath struct {
	Regexp
}

func (m *RegexpPath) Match(r *http.Request) bool {
	return m.MatchString(r.URL.Path)
}

// Extract returns positional and named variables extracted from the URL path.
func (m *RegexpPath) Extract(result *Result, r *http.Request) {
	result.Values = mergeValues(result.Values, m.Values(r.URL.Path))
}

// Build builds the URL path using the given positional and named variables,
// and writes it to the given URL.
func (m *RegexpPath) Build(u *url.URL, values url.Values) error {
	path, err := m.RevertValid(values)
	if err == nil {
		u.Path = path
	}
	return err
}
