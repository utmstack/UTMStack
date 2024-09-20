package utils

func PointerOf[t any](s t) *t{
	return &s
}
