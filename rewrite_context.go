package hblade

// RewriteContext is the interface for the URI rewrite ability.
type RewriteContext interface {
	Path() string
	SetPath(string)
}
