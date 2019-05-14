package main

import (
	"flag"
	"fmt"
	"os"
)

const msg = "Run the microservice and its dependencies with Docker"

type EnvDto struct {
	V           bool
	H           bool
	AppEnv      string
	ServiceName string
	Updated     string
	Ip          string
	MysqlPort   string

	RedisPort       string
	MongoPort       string
	SqlServerPort   string
	KafkaPort       string
	KafkaSecondPort string

	EventBrokerPort string
	NginxPort       string
	ZookeeperPort   string
}

var envDto = &EnvDto{}

func usage() {
	fmt.Fprintf(os.Stderr, `
Usage: run-test  [-s serviceName] [Options] 

Options:
`)
	flag.PrintDefaults()
}
func init() {
	flag.StringVar(&envDto.AppEnv, "appEnv", os.Getenv("appEnv"), "appEnv")
	flag.StringVar(&envDto.ServiceName, "s", os.Getenv("s"), "serviceName from mingbai")
	flag.StringVar(&envDto.Updated, "u", os.Getenv("u"), "remote or local")
	flag.StringVar(&envDto.Ip, "ip", os.Getenv("ip"), "ip")
	flag.StringVar(&envDto.MysqlPort, "mysqlPort", os.Getenv("mysqlPort"), "mysqlPort")

	flag.StringVar(&envDto.RedisPort, "redisPort", os.Getenv("redisPort"), "redisPort")
	flag.StringVar(&envDto.MongoPort, "mongoPort", os.Getenv("mongoPort"), "mongoPort")
	flag.StringVar(&envDto.SqlServerPort, "sqlServerPort", os.Getenv("sqlServerPort"), "sqlServerPort")
	flag.StringVar(&envDto.KafkaPort, "kafkaPort", os.Getenv("kafkaPort"), "kafkaPort")
	flag.StringVar(&envDto.KafkaSecondPort, "kafkaSecondPort", os.Getenv("kafkaSecondPort"), "kafkaSecondPort")

	flag.StringVar(&envDto.EventBrokerPort, "eventBrokerPort", os.Getenv("eventBrokerPort"), "eventBrokerPort")
	flag.StringVar(&envDto.NginxPort, "nginxPort", os.Getenv("nginxPort"), "nginxPort")
	flag.StringVar(&envDto.ZookeeperPort, "zookeeperPort", os.Getenv("zookeeperPort"), "zookeeperPort")
	flag.BoolVar(&envDto.H, "h", false, "help")
	flag.BoolVar(&envDto.V, "v", false, "version")

	flag.Usage = usage
}
