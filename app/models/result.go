package models

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(data interface{}) Result {
	return Result{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

func Fail(code int, msg string) Result {
	return Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

func Error(code int, msg string) Result {
	return Result{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}

func SuccessWithMsg(msg string) Result {
	return Result{
		Code: 200,
		Msg:  msg,
		Data: nil,
	}
}
