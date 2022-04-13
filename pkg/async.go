package pkg

type AsyncResult[T interface{}] struct {
	Err   error
	Value T
}
