package util

func MinInt(left, right int) int {
	if left > right {
		return right
	}

	return left
}
