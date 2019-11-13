package api

import "fmt"

func UnknownError(err error) Error {
	return newError(10001, "Unknown error", err.Error())
}
func InvalidParamError(field, condition string, err error) Error {
	return newError(10007, fmt.Sprintf("Invalid field(%s: %s)", field, condition), err.Error())
}
func ParameterParsingError(err error) Error {
	return newError(10008, "Parameter parsing error", err.Error())
}
func MissRequiredParamError(v string) Error {
	return newError(10009, fmt.Sprintf("'%s' is required parameter", v), nil)
}
func NotFoundError() Error {
	return newError(10010, "Resource is not found", nil)
}
func NotAuthorizedError() Error {
	return newError(10011, "Resource is not authorized", nil)
}
func NotAuthorizedActionError() Error {
	return newError(10012, "Action is not authorized", nil)
}
func StatusError(v string) Error {
	return newError(10013, fmt.Sprintf("'%s', Status not Allowed", v), nil)
}
func NotUpdatedError() Error {
	return newError(10014, "Resource is not updated", nil)
}
func NotDeletedError() Error {
	return newError(10015, "Resource is not deleted", nil)
}
func NotCreatedError() Error {
	return newError(10016, "Resource is not created", nil)
}

func InvalidFieldError(field string) Error {
	return newError(10018, fmt.Sprintf("Invalid fields [ %v ]", field), nil)
}

func newError(code int, message string, details interface{}) Error {
	return Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}
