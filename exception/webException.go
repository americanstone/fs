package exception

import "fmt"

type WebException struct {
	// 异常信息
	Message    string
	StatusCode int
	r          any
}

// ThrowWebException 抛出WebException异常
func ThrowWebException(statusCode int, err string) {
	panic(WebException{StatusCode: statusCode, Message: err})
}

// ThrowWebExceptionf 抛出WebException异常
func ThrowWebExceptionf(statusCode int, format string, a ...any) {
	panic(WebException{StatusCode: statusCode, Message: fmt.Sprintf(format, a...)})
}
