package file

// WithGHToken sets GitHub token.
func WithGHToken(token string) Option {
	return func(f *file) {
		f.GhToken = token
	}
}

// WithGHToken sets GitHub base URL.
// Useful if Enterprise Github is used on a custom domain.
func WithGHURL(url string) Option {
	return func(f *file) {
		f.GhUrl = url
	}
}

// WithDebug sets debug mode.
func WithDebug(debug bool) Option {
	return func(f *file) {
		f.Debug = debug
	}
}

// WithAbsPaths sets AbsPath mode.
func WithAbsPaths(b bool) Option {
	return func(f *file) {
		f.AbsPaths = b
	}
}

// WithStartDepth sets StartDepth.
func WithStartDepth(depth int) Option {
	return func(f *file) {
		f.StartDepth = depth
	}
}

// WithDepth sets Depth.
func WithDepth(depth int) Option {
	return func(f *file) {
		f.Depth = depth
	}
}

// WithEscape sets Escape mode.
func WithEscape(b bool) Option {
	return func(f *file) {
		f.Escape = b
	}
}

// WithIndent sets Indent mode.
func WithIndent(indent int) Option {
	return func(f *file) {
		f.Indent = indent
	}
}
