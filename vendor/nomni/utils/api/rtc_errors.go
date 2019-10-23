package api

import "fmt"

func RtcUnknownError() Error {
	return newError(81000, "系统错误，请与管理员联系.", nil)
}

func RtcServiceHasExistError() Error {
	return newError(81006, "微服务已经存在.", nil)
}

func RtcServiceHasNotExistError() Error {
	return newError(81007, "微服务不存在.", nil)
}

func RtcServiceHasNotFoundError() Error {
	return newError(81008, "没有查询到微服务.", nil)
}

func RtcServiceNotAllowDeleteError(ids []int) Error {
	return newError(81009, fmt.Sprintf("删除失败，有微服务 %v 依赖于这个服务.", ids), nil)
}
