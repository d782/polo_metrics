package utils

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Clamp[T Number](value T) T {

	if value < 0 {
		return 0
	}
	return value
}
