package conv

func Pointer[T any](v T) *T {
	return &v
}
