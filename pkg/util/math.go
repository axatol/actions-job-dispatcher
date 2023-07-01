package util

func MinInt(left, right int) int {
	if left > right {
		return right
	}

	return left
}

func MaxInt(left, right int) int {
	if left < right {
		return right
	}

	return left
}

func ClampInt(value, min, max int) int {
	value = MaxInt(value, min)
	value = MinInt(value, max)
	return value
}
