package toutbox

func Map[T any, U any](input []T, fn func(in T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = fn(v)
	}
	return result
}

func ref[T any](in T) *T { return &in }
