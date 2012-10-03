// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"net/http"
)

// All ------------------------------------------------------------------------

// NewAll returns a group of matchers that succeeds only if all of them match.
func NewAll(matchers []Matcher) All {
	return All(matchers)
}

// All is a set of matchers, and all of them must match.
type All []Matcher

func (m All) Match(r *http.Request) bool {
	for _, v := range m {
		if !v.Match(r) {
			return false
		}
	}
	return true
}

// One ------------------------------------------------------------------------

// NewOne returns a group of matchers that succeeds if one of them matches.
func NewOne(matchers []Matcher) One {
	return One(matchers)
}

// One is a set of matchers, and at least one of them must match.
type One []Matcher

func (m One) Match(r *http.Request) bool {
	for _, v := range m {
		if v.Match(r) {
			return true
		}
	}
	return false
}
