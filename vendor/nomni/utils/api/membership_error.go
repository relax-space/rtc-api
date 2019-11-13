package api

func MemberNotFoundError() Error {
	return newError(80001, "Member has not found", nil)
}
func MemberHasExistError() Error {
	return newError(80002, "Member has existed", nil)
}
func MemberMobileHasExistError() Error {
	return newError(80003, "Mobile has existed", nil)
}
func GradeRuleHasExistError() Error {
	return newError(80004, "GradeRule has existed", nil)
}
func GradeRuleNotFoundError() Error {
	return newError(80005, "GradeRule has not found", nil)
}
func MileageAddRuleHasExistError() Error {
	return newError(80006, "MileageAddRule has existed", nil)
}
func MileageAddRuleNotFoundError() Error {
	return newError(80007, "MileageAddRule has not found", nil)
}
func MileageUseRuleHasExistError() Error {
	return newError(80008, "MileageUseRule has existed", nil)
}
func MileageUseRuleNotFoundError() Error {
	return newError(80009, "MileageUseRule has not found", nil)
}
func MileageTypeNotFoundError() Error {
	return newError(80010, "MileageType has not found", nil)
}
func CurrentPointNotEnoughError() Error {
	return newError(80011, "CurrentPoint is not enough", nil)
}
func TradeNoHasExistError() Error {
	return newError(80012, "TradeNo has existed", nil)
}
func PreTradeNoHasExistError() Error {
	return newError(80013, "PreTradeNo has existed", nil)
}
func PreTradeNoNotFoundError() Error {
	return newError(80014, "PreTradeNo has not found", nil)
}
func SendSmsError() Error {
	return newError(80015, "SendSms error", nil)
}
func VerCodeNotRightError() Error {
	return newError(80016, "VerCode is not Right", nil)
}
func SyncFromCslError() Error {
	return newError(80017, "SyncFromCsl error", nil)
}
