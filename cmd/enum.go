package cmd

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

// YN
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

// ServiceType
type ServiceType int

const (
	EMPTYSERVER ServiceType = iota
	KAFKASERVER
	MYSQLSERVER
	SQLSERVERSERVER
	REDISSERVER
)

func (ServiceType) List() []string {
	return []string{"empty-server", "kafka", "mysql", "sqlserver", "redis"}
}

func (d ServiceType) String() string {
	return d.List()[d]
}

// BaseDbSource
type DbNet int

const (
	EMPTYDBNET DbNet = iota
	LOCALDBNET
	TCPDBNET
	HTTPDBNET
)

func (DbNet) List() []string {
	return []string{"empty", "local", "tcp", "http"}
}

func (d DbNet) String() string {
	return d.List()[d]
}
