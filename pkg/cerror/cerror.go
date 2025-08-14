package cerror

type CodeError struct {
	Code int32
	Msg  string
}

func NewCodeError(code int32, msg string) *CodeError {
	return &CodeError{
		Code: code,
		Msg:  msg,
	}
}
