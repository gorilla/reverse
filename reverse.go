// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reverse

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"regexp/syntax"
)

// Regexp stores a regular expression that can be "reverted" or "built":
// outermost capturing groups become placeholders to be filled by variables.
type Regexp struct {
	compiled *regexp.Regexp // compiled regular expression
	template string         // reverse template
	groups   []string       // order of positional and named capturing groups;
	// names for named and empty strings for positional
	indices []int // indices of the outermost groups
}

// CompileRegexp compiles a regular expression pattern and creates a template
// to revert it.
func CompileRegexp(pattern string) (*Regexp, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, err
	}
	tpl := &template{buffer: new(bytes.Buffer)}
	tpl.write(re)
	return &Regexp{
		compiled: compiled,
		template: tpl.buffer.String(),
		groups:   tpl.groups,
		indices:  tpl.indices,
	}, nil
}

// Compiled returns the compiled regular expression to be used for matching.
func (r *Regexp) Compiled() *regexp.Regexp {
	return r.compiled
}

// Template returns the reverse template for the regexp, in fmt syntax.
func (r *Regexp) Template() string {
	return r.template
}

// Groups returns an ordered list of the outermost capturing groups found in
// the regexp.
//
// Positional groups are listed as an empty string and named groups use
// the group name.
func (r *Regexp) Groups() []string {
	return r.groups
}

// Indices returns the indices of the outermost capturing groups found in
// the regexp.
//
// Not all indices may be present because nested capturing groups are ignored.
func (r *Regexp) Indices() []int {
	return r.indices
}

// Match returns whether the regexp matches the given string.
func (r *Regexp) MatchString(s string) bool {
	return r.compiled.MatchString(s)
}

// Values matches the regexp and returns the results for positional and
// named groups. Positional values are stored using an empty string as key.
// If the string doesn't match it returns nil.
func (r *Regexp) Values(s string) url.Values {
	match := r.compiled.FindStringSubmatch(s)
	if match != nil {
		values := url.Values{}
		for k, v := range r.groups {
			values.Add(v, match[r.indices[k]])
		}
		return values
	}
	return nil
}

// Revert builds a string for this regexp using the given values. Positional
// values use an empty string as key.
//
// The values are modified in place, and only the unused ones are left.
func (r *Regexp) Revert(values url.Values) (string, error) {
	vars := make([]interface{}, len(r.groups))
	for k, v := range r.groups {
		if len(values[v]) == 0 {
			return "", fmt.Errorf(
				"Missing key %q to revert the regexp "+
					"(expected a total of %d variables)", v, len(r.groups))
		}
		vars[k] = values[v][0]
		values[v] = values[v][1:]
	}
	return fmt.Sprintf(r.template, vars...), nil
}

// RevertValid is the same as Revert but it also validates the resulting
// string matching it against the compiled regexp.
//
// The values are modified in place, and only the unused ones are left.
func (r *Regexp) RevertValid(values url.Values) (string, error) {
	reverse, err := r.Revert(values)
	if err != nil {
		return "", err
	}
	if !r.compiled.MatchString(reverse) {
		return "", fmt.Errorf("Resulting string doesn't match the regexp: %q",
			reverse)
	}
	return reverse, nil
}

// template builds a reverse template for a regexp.
type template struct {
	buffer *bytes.Buffer
	groups []string // outermost capturing groups: empty string for
	// positional or name for named groups
	indices []int // indices of outermost capturing groups
	index   int   // current group index
	level   int   // current capturing group nesting level
}

// write writes a reverse template to the buffer.
func (t *template) write(re *syntax.Regexp) {
	switch re.Op {
	case syntax.OpLiteral:
		if t.level == 0 {
			for _, r := range re.Rune {
				t.buffer.WriteRune(r)
				if r == '%' {
					t.buffer.WriteRune('%')
				}
			}
		}
	case syntax.OpCapture:
		t.level++
		t.index++
		if t.level == 1 {
			t.groups = append(t.groups, re.Name)
			t.indices = append(t.indices, t.index)
			t.buffer.WriteString("%s")
		}
		for _, sub := range re.Sub {
			t.write(sub)
		}
		t.level--
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			t.write(sub)
		}
	}
}
