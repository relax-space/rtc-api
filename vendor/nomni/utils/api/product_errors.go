package api

func OutOfStockError() Error {
	return newError(21001, "Out of stock", nil)
}
func InvalidCodeError() Error {
	return newError(21002, "Code is invalid", nil)
}
