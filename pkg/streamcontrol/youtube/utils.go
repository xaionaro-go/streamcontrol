package youtube

func ptr[T any](in T) *T {
	return &in
}