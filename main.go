package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

const (
	TEMP_FILE        = "temp"
	EventBroker_Name = "event-broker-kafka"
)

const (
	PRIVATETOKEN         = "Su5_HzvQxtyANyDtzx_P"
	PreGitSshUrl         = "ssh://git@gitlab.p2shop.cn:822"
	PREGITHTTPURL        = "https://gitlab.p2shop.cn:8443"
	YMLNAMECONFIG        = "config"
	YMLNAMEDOCKERCOMPOSE = "docker-compose"
	REGISTRYNAME         = "registry.elandsystems.cn"
	PREWAIT              = "wait-"
	SUFSERVER            = "-server"
	PRETEST              = "test-"
)

var inPort = PortDto{
	Mysql:     "3306",
	Redis:     "6379",
	Mongo:     "27017",
	SqlServer: "1433",
	Kafka:     "9092",

	EventBroker: "3000",
	Nginx:       "80",
}

var outPort = PortDto{
	Kafka: "29092",
}

type PortDto struct {
	Mysql     string
	Redis     string
	Mongo     string
	SqlServer string
	Kafka     string

	EventBroker string
	Nginx       string
}
type ConfigDto struct {
	Scope   string
	Port    PortDto
	Project *ProjectDto
}
type ProjectDto struct {
	IsMulti          bool     //a git contains multiple microservices
	ServiceName      string   //eg. ipay-api
	GitShortPath     string   //eg. ipay/ipay-api
	Envs             []string // from jenkins
	IsProjectKafka   bool
	Ports            []string
	Databases        []string //mysql,redis,mongo,sqlserver
	StreamNames      []string
	ParentFolderName string

	GitRaw      string
	SubProjects []*ProjectDto
}

func main() {

	c, err := LoadEnv()
	if err != nil {
		fmt.Println(err)
		return
	}

	//1.download sql data
	if shouldUpdateData(c.Scope) {
		if err := deleteFileRegex(TEMP_FILE + "/*.sql"); err != nil {
			fmt.Println(err)
			return
		}
		if err := fetchsqlTofile(c.Project); err != nil {
			fmt.Println(err)
			return
		}
	}

	//2. generate docker-compose
	if shouldUpdateCompose(c.Scope) {
		viper := viper.New()
		compose := Compose{}
		if shouldStartKakfa(c.Project) {
			compose.setComposeKafka(viper, c.Port.Kafka)
		}
		if shouldStartMysql(c.Project) {
			compose.setComposeMysql(viper, c.Port.Mysql)
		}
		if shouldStartRedis(c.Project) {
			compose.setComposeRedis(viper, c.Port.Redis)
		}

		if shouldStartEventBroker(c.Project) {
			streamName := streamList(c.Project)
			EventBroker{}.SetEventBroker(viper, c.Port.EventBroker, streamName)
		}
		compose.setComposeApp(viper, c.Project)
		compose.setComposeNginx(viper, c.Project.ServiceName, c.Port.Nginx)
		ComposeWait{}.setWaitCompose(viper, c.Project)

		// if err := writeConfig(TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", viper); err != nil {
		// 	fmt.Printf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		// 	return
		// }

		ymlStr, err := yamlStringSettings(viper)
		if err != nil {
			fmt.Printf("write to %v error:%v\n", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		}

		ymlStr = strings.Replace(ymlStr, "kafka_advertised_listeners", "KAFKA_ADVERTISED_LISTENERS", -1)
		ymlStr = strings.Replace(ymlStr, "kafka_inter_broker_listener_name", "KAFKA_INTER_BROKER_LISTENER_NAME", -1)
		ymlStr = strings.Replace(ymlStr, "kafka_listener_security_protocol_map", "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", -1)
		ymlStr = strings.Replace(ymlStr, "kafka_listeners", "KAFKA_LISTENERS", -1)
		ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)

		if writeFile(TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", ymlStr); err != nil {
			fmt.Printf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		}
		//writeFile
	}

	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	//3. run docker-compose
	if shouldRestartData(c.Scope) {
		//delete volume
		if _, err = Cmd("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose downed!")
	}
	if _, err = Cmd("docker-compose", "-f", dockercompose, "pull"); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	fmt.Println("==> compose pulled!")

	if shouldRestartApp(c.Scope) {
		if _, err = Cmd("docker-compose", "-f", dockercompose, "build"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose builded!")
	}

	go func() {
		if _, err = Cmd("docker-compose", "-f", dockercompose, "up"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)
	go func() {
		for s := range signals {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				os.Exit(0)
			}
		}
	}()
	time.Sleep(10 * time.Second)
	fmt.Println("==> compose may have been up!")
	time.Sleep(10 * time.Minute)
}
