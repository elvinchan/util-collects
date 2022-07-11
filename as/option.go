package as

type Option func(*As)

// WithFailDirectly specified fail directly for checks of as object.
func WithFailDirectly(b bool) Option {
	return func(a *As) {
		a.failDirectly = b
	}
}
