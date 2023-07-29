# reverse

![testing](https://github.com/gorilla/reverse/actions/workflows/test.yml/badge.svg)
[![codecov](https://codecov.io/github/gorilla/reverse/branch/main/graph/badge.svg)](https://codecov.io/github/gorilla/reverse)
[![godoc](https://godoc.org/github.com/gorilla/reverse?status.svg)](https://godoc.org/github.com/gorilla/reverse)
[![sourcegraph](https://sourcegraph.com/github.com/gorilla/reverse/-/badge.svg)](https://sourcegraph.com/github.com/gorilla/reverse?badge)

![Gorilla Logo](https://github.com/gorilla/.github/assets/53367916/d92caabf-98e0-473e-bfbf-ab554ba435e5)

Package gorilla/reverse is a set of utilities to create request routers.

It provides interfaces to match and extract variables from an HTTP request
and build URLs for registered routes. It also has a variety of matcher
implementations for all kinds of request attributes, among other utilities.

For example, the Regexp type produces reversible regular expressions that
can be used to generate URLs for a regexp-based mux. To demonstrate, let's
compile a simple regexp:

```go
regexp, err := reverse.CompileRegexp(`/foo/1(\d+)3`)
```

Now we can call regexp.Revert() passing variables to fill the capturing groups.
Because our variable is not named, we use an empty string as key for
url.Values, like this:

```go
// url is "/foo/123".
url, err := regexp.Revert(url.Values{"": {"2"}})
```

Non-capturing groups are ignored, but named capturing groups can be filled
normally. Just set the key in url.Values:

```go
regexp, err := reverse.CompileRegexp(`/foo/1(?P<two>\d+)3`)
if err != nil {
    panic(err)
}
// url is "/foo/123".
url, err := re.Revert(url.Values{"two": {"2"}})
```

There are a few limitations that can't be changed:

1. Nested capturing groups are ignored; only the outermost groups become
a placeholder. So in `1(\d+([a-z]+))3` there is only one placeholder
although there are two capturing groups: re.Revert(url.Values{"": {"2", "a"}})
results in "123" and not "12a3".
2. Literals inside capturing groups are ignored; the whole group becomes
a placeholder.
