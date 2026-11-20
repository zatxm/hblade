package hblade

import (
	"slices"
	"strings"
	"testing"
)

func TestRouterFind(t *testing.T) {
	tests := []struct {
		routes []string
		path   string
		want   string // matched route, "" = not found
		params string // "k=v,k=v"
	}{
		// Backtracking from a partially matching static branch.
		{[]string{"/user/:id", "/user/new"}, "/user/newx", "/user/:id", "id=newx"},
		{[]string{"/user/:id", "/user/new"}, "/user/ne", "/user/:id", "id=ne"},
		{[]string{"/a/:x/d", "/a/b/c"}, "/a/b/d", "/a/:x/d", "x=b"},
		{[]string{"/a/:x/d", "/a/bc"}, "/a/b/d", "/a/:x/d", "x=b"},
		{[]string{"/a/:x/:y/z", "/a/b/c"}, "/a/b/q/z", "/a/:x/:y/z", "x=b,y=q"},
		{[]string{"/s/*f", "/s/:x/a"}, "/s/q/b", "/s/*f", "f=q/b"},

		// Existing behavior.
		{[]string{"/user/:id", "/user/new"}, "/user/new", "/user/new", ""},
		{[]string{"/a/:x/d", "/a/b/c"}, "/a/b/c", "/a/b/c", ""},
		{[]string{"/u/:id", "/u/:id/p/:pid"}, "/u/1/p/2", "/u/:id/p/:pid", "id=1,pid=2"},
		{[]string{"/u/:id"}, "/u/1/", "/u/:id", "id=1"},
		{[]string{"/u/:id"}, "/u/1/extra", "", ""},
		{[]string{"/u/:id"}, "/u/", "", ""},
		{[]string{"/blog"}, "/blog/", "/blog", ""},
		{[]string{"/a/*f", "/a/b/c"}, "/a/b/x", "/a/*f", "f=b/x"},
		{[]string{"/*f", "/a/b"}, "/a/c", "/*f", "f=a/c"},
		{[]string{"/ab", "/a/:x"}, "/a/1", "/a/:x", "x=1"},
		{[]string{"/", "/a"}, "/", "/", ""},
		{[]string{"/a"}, "/b", "", ""},
	}
	for _, tt := range tests {
		r := &Router[string]{}
		for _, route := range tt.routes {
			r.Add("GET", route, route)
		}
		var ps Params
		got := r.Find("GET", tt.path, &ps)
		var params []string
		for i := range ps.Len() {
			k, v := ps.At(i)
			params = append(params, k+"="+v)
		}
		if got != tt.want || strings.Join(params, ",") != tt.params {
			t.Errorf("routes %v, GET %s = %q [%s], want %q [%s]",
				tt.routes, tt.path, got, strings.Join(params, ","), tt.want, tt.params)
		}
	}
}

func TestRouterLookupCallback(t *testing.T) {
	r := &Router[string]{}
	r.Add("GET", "/u/:id/p/:pid", "ok")
	var got []string
	if r.Lookup("GET", "/u/1/p/2", func(k, v string) { got = append(got, k+"="+v) }) != "ok" ||
		!slices.Equal(got, []string{"id=1", "pid=2"}) {
		t.Fatalf("Lookup params = %v", got)
	}
}

func TestRouterNameConflict(t *testing.T) {
	tests := []struct {
		routes []string
		panics bool
	}{
		{[]string{"/a/:x", "/a/:y"}, true},
		{[]string{"/a/:x/b", "/a/:y/c"}, true},
		{[]string{"/a/:x", "/a/:x/b", "/a/:x/c"}, false},
		{[]string{"/a/:x", "/a/:x"}, false},
		{[]string{"/s/*f", "/s/*g"}, true},
		{[]string{"/s/*f", "/s/*f"}, false},
		{[]string{"/s/*f", "/s/:x/a"}, false},
		{[]string{"/a/:x", "/b/:y"}, false},
	}
	for _, tt := range tests {
		func() {
			defer func() {
				if got := recover() != nil; got != tt.panics {
					t.Errorf("routes %v: panicked = %v, want %v", tt.routes, got, tt.panics)
				}
			}()
			r := &Router[string]{}
			for _, route := range tt.routes {
				r.Add("GET", route, route)
			}
		}()
	}
}
