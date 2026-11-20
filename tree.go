package hblade

import "strings"

// Tree represents a radix tree.
type Tree[T any] struct {
	root treeNode[T]
}

// Add adds a new element to the tree.
func (tree *Tree[T]) Add(path string, data T) {
	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0
	node := &tree.root

	for {
	begin:
		switch node.kind {
		case parameter:
			// This only occurs when the same parameter based route is added twice.
			// node: /post/:id|
			// path: /post/:id|
			if i == len(path) {
				node.setData(data)
				return
			}

			// When we hit a separator, we'll search for a fitting child.
			if path[i] == separator {
				node, offset, _ = node.end(path, data, i, offset)
				goto next
			}

		default:
			if i == len(path) {
				// The path already exists.
				// node: /blog|
				// path: /blog|
				if i-offset == len(node.prefix) {
					node.setData(data)
					return
				}

				// The path ended but the node prefix is longer.
				// node: /blog|feed
				// path: /blog|
				node.split(i-offset, "", data)
				return
			}

			// The node we just checked is entirely included in our path.
			// node: /|
			// path: /|blog
			if i-offset == len(node.prefix) {
				var control flow
				node, offset, control = node.end(path, data, i, offset)

				switch control {
				case flowStop:
					return
				case flowBegin:
					goto begin
				case flowNext:
					goto next
				}
			}

			// We got a conflict.
			// node: /b|ag
			// path: /b|riefcase
			if path[i] != node.prefix[i-offset] {
				node.split(i-offset, path[i:], data)
				return
			}
		}

	next:
		i++
	}
}

// Params holds route parameters collected by a lookup without allocating.
type Params struct {
	keys   [maxParams]string
	values [maxParams]string
	n      int
}

// Get returns the value of the named parameter, or "".
func (ps *Params) Get(key string) string {
	for i := range ps.n {
		if ps.keys[i] == key {
			return ps.values[i]
		}
	}
	return ""
}

// Len returns the number of parameters.
func (ps *Params) Len() int {
	return ps.n
}

// At returns the i-th parameter.
func (ps *Params) At(i int) (key, value string) {
	return ps.keys[i], ps.values[i]
}

// Reset clears the parameters, dropping references to request strings.
func (ps *Params) Reset() {
	ps.truncate(0)
}

func (ps *Params) add(key, value string) {
	if ps.n < maxParams {
		ps.keys[ps.n] = key
		ps.values[ps.n] = value
		ps.n++
	}
}

func (ps *Params) truncate(n int) {
	// Routes have few params, so an inline loop beats clear()'s memclr call.
	for i := n; i < ps.n; i++ {
		ps.values[i] = ""
	}
	ps.n = n
}

// backtrack is a pending alternative at a node whose static child was tried first.
type backtrack[T any] struct {
	node  *treeNode[T]
	path  string
	n     int  // params count to restore
	param bool // resume with the parameter child, otherwise the wildcard
}

// Find returns the data for path and appends its parameters to ps, without
// allocating. Matching priority per segment is static > parameter > wildcard,
// backtracking when a more specific branch doesn't lead to a route.
//
// It first descends greedily, which is the first path a depth-first search would
// take, so a hit here equals the full search result. Only a miss falls back to
// the complete search in findBacktrack.
func (tree *Tree[T]) Find(path string, ps *Params) T {
	var empty T
	node := &tree.root
	if p := node.prefix; p != "" && (path == "" || path[0] != p[0] || !hasPrefixAfterFirst(path, p)) {
		return empty
	}
	rest := path[len(node.prefix):]
	base := ps.n

	for rest != "" {
		if child := node.child(rest[0]); child != nil && hasPrefixAfterFirst(rest, child.prefix) {
			node = child
			rest = rest[len(child.prefix):]
			continue
		}

		if param := node.parameter; param != nil {
			if end := segmentEnd(rest); end > 0 {
				ps.add(param.prefix, rest[:end])
				node = param
				rest = rest[end:]
				continue
			}
		}

		if wild := node.wildcard; wild != nil && wild.hasData {
			ps.add(wild.prefix, rest)
			return wild.data
		}
		goto miss
	}
	if node.hasData {
		return node.data
	}

miss:
	ps.truncate(base)
	return tree.findBacktrack(path, ps)
}

// segmentEnd returns the index of the first separator in s, or len(s).
// Segments are short, so an inline loop beats strings.IndexByte.
func segmentEnd(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == separator {
			return i
		}
	}
	return len(s)
}

// hasPrefixAfterFirst reports whether s starts with prefix, assuming their first
// bytes are already known to be equal. Route prefixes are short, so an inline
// loop beats a memequal call.
func hasPrefixAfterFirst(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	s = s[:len(prefix)]
	for i := 1; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

// findBacktrack is the complete depth-first search used when the greedy
// descent in Find failed after passing an untried alternative.
func (tree *Tree[T]) findBacktrack(path string, ps *Params) T {
	var empty T
	node := &tree.root
	if !strings.HasPrefix(path, node.prefix) {
		return empty
	}
	path = path[len(node.prefix):]

	// Alternatives are only pushed at nodes that have them, so this rarely grows
	// beyond the on-stack buffer; deeper backtracking spills to the heap.
	var buf [8]backtrack[T]
	stack := buf[:0]
	base := ps.n

	for {
		// Here node's prefix has been consumed and path is the remainder.
		if path == "" {
			if node.hasData {
				return node.data
			}
			goto fail
		}

		if child := node.child(path[0]); child != nil {
			if hasPrefixAfterFirst(path, child.prefix) {
				if node.parameter != nil || node.wildcard != nil {
					stack = append(stack, backtrack[T]{node, path, ps.n, true})
				}
				node = child
				path = path[len(child.prefix):]
				continue
			}
		}

	tryParam:
		if param := node.parameter; param != nil {
			if end := segmentEnd(path); end > 0 {
				if node.wildcard != nil {
					stack = append(stack, backtrack[T]{node, path, ps.n, false})
				}
				ps.add(param.prefix, path[:end])
				node = param
				path = path[end:]
				continue
			}
		}

	tryWildcard:
		if wild := node.wildcard; wild != nil && wild.hasData {
			ps.add(wild.prefix, path)
			return wild.data
		}

	fail:
		if len(stack) == 0 {
			ps.truncate(base)
			return empty
		}
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		ps.truncate(f.n)
		node, path = f.node, f.path
		if f.param {
			goto tryParam
		}
		goto tryWildcard
	}
}

// Lookup finds the data for the given path and reports each parameter via addParameter.
func (tree *Tree[T]) Lookup(path string, addParameter func(key string, value string)) T {
	var ps Params
	data := tree.Find(path, &ps)
	for i := range ps.n {
		addParameter(ps.keys[i], ps.values[i])
	}
	return data
}
