package httpx

// Standard API response wrapper
// code: 0 success; non-zero for errors
// message: human readable
// data: optional payload

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *T     `json:"data,omitempty"`
}

func Ok[T any](data *T) Response[T] {
	return Response[T]{Code: 0, Message: "OK", Data: data}
}

func OkMsg[T any](msg string, data *T) Response[T] {
	return Response[T]{Code: 0, Message: msg, Data: data}
}

func Err[T any](code int, msg string) Response[T] {
	return Response[T]{Code: code, Message: msg, Data: nil}
}
