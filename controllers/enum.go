package controllers

// DateBaseType
type DateBaseType int

const (
	MYSQL DateBaseType = iota
	REDIS
	MONGO
	SQLSERVER
)

func (DateBaseType) List() []string {
	return []string{"mysql", "redis", "mongo", "sqlserver"}
}

func (d DateBaseType) String() string {
	return d.List()[d]
}
