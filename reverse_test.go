// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"net/url"
	"testing"
)

func copyValues(values url.Values) url.Values {
	rv := url.Values{}
	for k, v := range values {
		rv[k] = make([]string, len(v))
		copy(rv[k], v)
	}
	return rv
}

type reverseTest struct {
	pattern string
	values  url.Values
	result  string
	valid   bool
}

var reverseTests = []reverseTest{
	{
		pattern: `^1(\d+)3$`,
		values:  url.Values{"": []string{"2"}},
		result:  "123",
		valid:   true,
	},
	{
		pattern: `^1(\d+)3$`,
		values:  url.Values{"": []string{"a"}},
		result:  "1a3",
		valid:   false,
	},
	{
		pattern: `^4(?P<foo>\d+)6$`,
		values:  url.Values{"foo": []string{"5"}},
		result:  "456",
		valid:   true,
	},
	{
		pattern: `^4(?P<foo>\d+)6$`,
		values:  url.Values{"foo": []string{"b"}},
		result:  "4b6",
		valid:   false,
	},
	{
		pattern: `^7(?P<foo>\d)(\d)0$`,
		values:  url.Values{"": []string{"9"}, "foo": []string{"8"}},
		result:  "7890",
		valid:   true,
	},
	{
		pattern: `^7(?P<foo>\d)(\d)0$`,
		values:  url.Values{"": []string{"d"}, "foo": []string{"c"}},
		result:  "7cd0",
		valid:   false,
	},
	{
		pattern: `(?P<foo>\d)(\d)(?P<foo>\d)`,
		values:  url.Values{"": []string{"2"}, "foo": []string{"1", "3"}},
		result:  "123",
		valid:   true,
	},
	{
		pattern: `(?P<foo>\d)(\d)(?P<foo>\d)`,
		values:  url.Values{"": []string{"b"}, "foo": []string{"a", "c"}},
		result:  "abc",
		valid:   false,
	},
}

func TestReverseRegexp(t *testing.T) {
	for _, test := range reverseTests {
		r, err := CompileRegexp(test.pattern)
		if err != nil {
			t.Fatal(err)
		}
		// MatchString()
		if r.MatchString(test.result) != test.valid {
			t.Errorf("%q: expected match %q, got %q", test.pattern, test.valid, !test.valid)
		}
		// Values()
		if test.valid {
			values := r.Values(test.result)
			reverted, err := r.Revert(copyValues(values))
			if err != nil {
				t.Fatalf("%s: pattern: %q, values: %#v, indices: %#v, groups: %#v", err, test.pattern, values, r.indices, r.groups)
			}
			if reverted != test.result {
				t.Errorf("%q: expected reverted %q, got %q for values %v", test.pattern, test.result, reverted, values)
			}
		}
		// Revert()
		reverted, err := r.Revert(copyValues(test.values))
		if err != nil {
			t.Fatalf("%s: pattern: %q, values: %#v, indices: %#v, groups: %#v", err, test.pattern, test.values, r.indices, r.groups)
		}
		if reverted != test.result {
			t.Errorf("%q: expected reverted %q, got %q for values %v", test.pattern, test.result, reverted, test.values)
		}
		// RevertValid()
		reverted, err = r.RevertValid(copyValues(test.values))
		if test.valid {
			if err != nil {
				t.Errorf("%q: expected success on RevertValid, got %v", test.pattern, err)
			} else if reverted != test.result {
				t.Errorf("%q: expected reverted %q, got %q for values %v", test.pattern, test.result, reverted, test.values)
			}
		} else {
			if err == nil {
				t.Errorf("%q: expected error on RevertValid", test.pattern)
			}
		}
	}
}

type groupTest struct {
	pattern string
	groups  []string
	indices []int
}

var groupTests = []groupTest{
	groupTest{
		pattern: `^1(\d+)3$`,
		groups:  []string{""},
		indices: []int{1},
	},
	groupTest{
		pattern: `^1(\d+([a-z]+)(\d+([a-z]+)))(?P<foo>\d+)3([a-z]+(\d+))(?P<bar>\d+)$`,
		groups:  []string{"", "foo", "", "bar"},
		indices: []int{1, 5, 6, 8},
	},
}

func TestGroups(t *testing.T) {
	for _, test := range groupTests {
		r, err := CompileRegexp(test.pattern)
		if err != nil {
			t.Fatal(err)
		}
		groups := r.Groups()
		indices := r.Indices()
		if !stringSliceEqual(test.groups, groups) {
			t.Errorf("Expected %v, got %v", test.groups, groups)
		}
		if !intSliceEqual(test.indices, indices) {
			t.Errorf("Expected %v, got %v", test.indices, indices)
		}
	}
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}
