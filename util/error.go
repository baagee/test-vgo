package util

//自定义错误接口
type Error interface {
	//继承error
	error
	// 定义状态码
	Status() int
}

// 定义自定义的错误，可以指定错误码
type StatusError struct {
	//错误码
	Code int
	//错误信息
	Err error
}

//实现 error接口
func (s StatusError) Error() string {
	return s.Err.Error()
}

//实现 StatusError接口
func (s StatusError) Status() int {
	return s.Code
}
