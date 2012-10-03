// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"net/http"
	"net/url"
	"testing"
)

func equalStringSlice(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for k, v := range s1 {
		if s2[k] != v {
			return false
		}
	}
	return true
}

func equalValues(u1, u2 url.Values) bool {
	if len(u1) != len(u2) {
		return false
	}
	for k, v := range u1 {
		if !equalStringSlice(v, u2[k]) {
			return false
		}
	}
	return true
}

func testMatcher(t *testing.T, name string, m Matcher, r *http.Request, expect bool) {
	result := m.Match(r)
	if result != expect {
		t.Errorf("%s: got %v, expected %v", name, result, expect)
	}
}

func TestHost(t *testing.T) {
	const name = "Host"
	type test struct {
		host   string
		rURL   string
		expect bool
	}
	tests := []test{
		{"domain.com", "http://domain.com", true},
		{"domain.com", "http://other.com", false},
	}
	for _, v := range tests {
		r, err := http.NewRequest("GET", v.rURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, name, NewHost(v.host), r, v.expect)
	}
}

func TestMethod(t *testing.T) {
	const name = "Method"
	type test struct {
		methods []string
		rMethod string
		expect  bool
	}
	tests := []test{
		{[]string{"GET", "POST"}, "GET", true},
		{[]string{"GET", "POST"}, "POST", true},
		{[]string{"get", "post"}, "GET", true},
		{[]string{"get", "post"}, "POST", true},
		{[]string{"POST", "PUT"}, "GET", false},
	}
	for _, v := range tests {
		r, err := http.NewRequest(v.rMethod, "http://domain.com", nil)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, name, NewMethod(v.methods), r, v.expect)
	}
}

func TestRegexpHost(t *testing.T) {
	const name = "RegexpHost"
	type test struct {
		host   string
		rURL   string
		expect bool
		values url.Values
	}
	tests := []test{
		{`(?P<subdomain>[a-z]+)\.domain\.com`, "http://sub.domain.com", true, url.Values{"subdomain": {"sub"}}},
		{`(?P<subdomain>[a-z]+)\.domain\.com`, "http://123.domain.com", false, nil},
	}
	for _, v := range tests {
		r, err := http.NewRequest("GET", v.rURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		matcher, err := NewRegexpHost(v.host)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, name, matcher, r, v.expect)
		result := Result{}
		matcher.Extract(&result, r)
		if v.expect {
			if !equalValues(v.values, result.Values) {
				t.Errorf("%s: expected %v, got %v", name, v.values, result.Values)
			}
			u := url.URL{}
			if err := matcher.Build(&u, result.Values); err != nil {
				t.Errorf("%s: error building URL", name)
			} else {
				u2, _ := url.Parse(v.rURL)
				if u.Host != u2.Host {
					t.Errorf("%s: expected %q, got %q", name, u2.Host, u.Host)
				}
			}
		}
	}
}

func TestRegexpPath(t *testing.T) {
	const name = "RegexpPath"
	type test struct {
		path   string
		rURL   string
		expect bool
		values url.Values
	}
	tests := []test{
		{`/(?P<abc>[a-z]+)/(?P<ghi>[a-z]+)`, "http://domain.com/def/jkl", true, url.Values{"abc": {"def"}, "ghi": {"jkl"}}},
		{`/(?P<abc>[a-z]+)/(?P<ghi>[a-z]+)`, "http://domain.com/123/456", false, nil},
	}
	for _, v := range tests {
		r, err := http.NewRequest("GET", v.rURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		matcher, err := NewRegexpPath(v.path)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, name, matcher, r, v.expect)
		result := Result{}
		matcher.Extract(&result, r)
		if v.expect {
			if !equalValues(v.values, result.Values) {
				t.Errorf("%s: expected %v, got %v", name, v.values, result.Values)
			}
			u := url.URL{}
			if err := matcher.Build(&u, result.Values); err != nil {
				t.Errorf("%s: error building URL", name)
			} else {
				u2, _ := url.Parse(v.rURL)
				if u.Path != u2.Path {
					t.Errorf("%s: expected %q, got %q", name, u2.Path, u.Path)
				}
			}
		}
	}
}

func TestScheme(t *testing.T) {
	const name = "Scheme"
	type test struct {
		schemes []string
		rURL    string
		expect  bool
	}
	tests := []test{
		{[]string{"http", "https"}, "http://domain.com", true},
		{[]string{"http", "https"}, "https://domain.com", true},
		{[]string{"https"}, "http://domain.com", false},
	}
	for _, v := range tests {
		r, err := http.NewRequest("GET", v.rURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		testMatcher(t, name, NewScheme(v.schemes), r, v.expect)
	}
}
