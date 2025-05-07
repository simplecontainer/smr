package helpers

func RemoveElement[T any](slice []T, k int) []T {
	if k < 0 || k >= len(slice) {
		return slice
	}

	return append(slice[:k], slice[k+1:]...)
}
