package api

func RedundantItemsError(err error) Error {
	return newError(22001, "Redundant items", err.Error())
}
