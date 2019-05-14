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
	flag.StringVar(&envDto.AppEnv, "appEnv", os.Getenv("appEnv"), "appEnv [qa prd]")
	flag.StringVar(&envDto.ServiceName, "s", os.Getenv("s"), "serviceName from mingbai")
	flag.StringVar(&envDto.Updated, "u", os.Getenv("u"), "data source from [remote local]")
	flag.StringVar(&envDto.Ip, "ip", os.Getenv("ip"), "ip format:10.202.101.43")
	flag.StringVar(&envDto.MysqlPort, "mysqlPort", os.Getenv("mysqlPort"), "mysqlPort"+inPort.Mysql)

	flag.StringVar(&envDto.RedisPort, "redisPort", os.Getenv("redisPort"), "redisPort"+inPort.Redis)
	flag.StringVar(&envDto.MongoPort, "mongoPort", os.Getenv("mongoPort"), "mongoPort"+inPort.Mongo)
	flag.StringVar(&envDto.SqlServerPort, "sqlServerPort", os.Getenv("sqlServerPort"), "sqlServerPort"+inPort.SqlServer)
	flag.StringVar(&envDto.KafkaPort, "kafkaPort", os.Getenv("kafkaPort"), "kafkaPort"+inPort.Kafka)
	flag.StringVar(&envDto.KafkaSecondPort, "kafkaSecondPort", os.Getenv("kafkaSecondPort"), "kafkaSecondPort"+inPort.KafkaSecond)

	flag.StringVar(&envDto.EventBrokerPort, "eventBrokerPort", os.Getenv("eventBrokerPort"), "eventBrokerPort"+inPort.EventBroker)
	flag.StringVar(&envDto.NginxPort, "nginxPort", os.Getenv("nginxPort"), "nginxPort"+inPort.Nginx)
	flag.StringVar(&envDto.ZookeeperPort, "zookeeperPort", os.Getenv("zookeeperPort"), "zookeeperPort"+inPort.Zookeeper)
	flag.BoolVar(&envDto.H, "h", false, "help")
	flag.BoolVar(&envDto.V, "v", false, "version")

	flag.Usage = usage
}
