package rtmp

// ReadEmptyError 读入字节错误
type ReadEmptyError struct {
}

func (err ReadEmptyError) Error() string {
	return "Read buffer empty"
}
