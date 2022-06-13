package hblade

import "strings"

// controlFlow tells the main loop what it should do next.
type controlFlow int

// controlFlow values.
const (
	controlStop controlFlow = iota
	controlBegin
	controlNext
)

// dataType specifies which type of data we are going to save for each node.
type dataType = Handler

// tree represents a radix tree.
type tree struct {
	root        treeNode
	static      map[string]dataType
	mcheck      map[string]int
	canBeStatic [2048]bool
}

// add adds a new element to the tree.
func (tree *tree) add(path string, data dataType) {
	if !strings.Contains(path, ":") && !strings.Contains(path, "*") {
		if tree.static == nil {
			tree.static = map[string]dataType{}
			tree.mcheck = map[string]int{}
		}

		tree.static[path] = data
		tree.mcheck[path] = 0
		tree.canBeStatic[len(path)] = true
		return
	}

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
				node.data = data
				return
			}

			// When we hit a separator, we'll search for a fitting child.
			if path[i] == separator {
				var control controlFlow
				node, offset, control = node.end(path, data, i, offset)

				switch control {
				case controlStop:
					return
				case controlBegin:
					goto begin
				case controlNext:
					goto next
				}
			}

		default:
			if i == len(path) {
				// The path already exists.
				// node: /blog|
				// path: /blog|
				if i-offset == len(node.prefix) {
					node.data = data
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
				var control controlFlow
				node, offset, control = node.end(path, data, i, offset)

				switch control {
				case controlStop:
					return
				case controlBegin:
					goto begin
				case controlNext:
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

// find finds the data for the given path and assigns it to c.handler, if available.
func (tree *tree) find(path string, c *Context) {
	if tree.canBeStatic[len(path)] {
		handler, found := tree.static[path]

		if found {
			c.handler = handler
			return
		}
	}

	var (
		i                  uint
		offset             uint
		lastWildcardOffset uint
		lastWildcard       *treeNode
		node               = &tree.root
	)

begin:
	// Search tree for equal parts until we can no longer proceed
	for {
		// We reached the end.
		if i == uint(len(path)) {
			// node: /blog|
			// path: /blog|
			if i-offset == uint(len(node.prefix)) {
				c.handler = node.data
				return
			}

			// node: /blog|feed
			// path: /blog|
			c.handler = nil
			return
		}

		// The node we just checked is entirely included in our path.
		// node: /|
		// path: /|blog
		if i-offset == uint(len(node.prefix)) {
			if node.wildcard != nil {
				lastWildcard = node.wildcard
				lastWildcardOffset = i
			}

			char := path[i]

			if char >= node.startIndex && char < node.endIndex {
				index := node.indices[char-node.startIndex]

				if index != 0 {
					node = node.children[index]
					offset = i
					i++
					continue
				}
			}

			// node: /|:id
			// path: /|blog
			if node.parameter != nil {
				node = node.parameter
				offset = i
				i++

				for {
					// We reached the end.
					if i == uint(len(path)) {
						c.addParameter(node.prefix, path[offset:i])
						c.handler = node.data
						return
					}

					// node: /:id|/posts
					// path: /123|/posts
					if path[i] == separator {
						c.addParameter(node.prefix, path[offset:i])
						index := node.indices[separator-node.startIndex]
						node = node.children[index]
						offset = i
						i++
						goto begin
					}

					i++
				}
			}

			// node: /|*any
			// path: /|image.png
			if node.wildcard != nil {
				c.addParameter(node.wildcard.prefix, path[i:])
				c.handler = node.wildcard.data
				return
			}

			c.handler = nil
			return
		}

		// We got a conflict.
		// node: /b|ag
		// path: /b|riefcase
		if path[i] != node.prefix[i-offset] {
			if lastWildcard != nil {
				c.addParameter(lastWildcard.prefix, path[lastWildcardOffset:])
				c.handler = lastWildcard.data
				return
			}

			c.handler = nil
			return
		}

		i++
	}
}

// bind binds all handlers to a new one provided by the callback.
func (tree *tree) bind(transform func(Handler) Handler) {
	tree.root.each(func(node *treeNode) {
		if node.data != nil && node.gocheck != 1 {
			node.data = transform(node.data)
			node.gocheck = 1
		}
	})

	for key := range tree.static {
		if tree.mcheck[key] == 0 {
			value := tree.static[key]
			tree.static[key] = transform(value)
			tree.mcheck[key] = 1
		}
	}
}
