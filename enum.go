package main

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

// ScopeType
type ScopeType int

const (
	REMOTE ScopeType = iota
	LOCAL
)

func (ScopeType) List() []string {
	return []string{"remote", "local"}
}

func (d ScopeType) String() string {
	return d.List()[d]
}

// ScopeType
type YN int

const (
	N YN = iota
	Y
)

func (YN) List() []string {
	return []string{"n", "y"}
}

func (d YN) String() string {
	return d.List()[d]
}
