package binancews

type Stream[T any] struct {
	Stream string `json:"stream"`
	Data   T      `json:"data"`
}
