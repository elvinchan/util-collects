package as

// FailDirectly specified fail directly for checks of as object.
func (a *As) FailDirectly(b bool) *As {
	a.failDirectly.Store(b)
	return a
}
