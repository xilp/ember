package measure

func Max(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

func Min(a, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}

func Count(a, b int64) int64 {
	return a + 1
}

func Sum(a, b int64) int64 {
	return a + b
}
