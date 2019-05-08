package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	mysql "github.com/go-sql-driver/mysql"

	kafkautil "github.com/segmentio/kafka-go"

	"github.com/spf13/viper"
)

var (
	app_env = ""
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
	REGISTRYELAND        = "registry.elandsystems.cn"
	PREWAIT              = "wait-"
	SUFSERVER            = "-server"
	PRETEST              = "test-"
	P2SHOPHOST           = "https://gateway.p2shop.com.cn"
	REGISTRYQA           = "registry.p2shop.com.cn"
	WAITIMAGE            = "waisbrot/wait" //xiaoxinmiao/wait:0.0.2
)

var inPort = PortDto{
	Mysql:     "3306",
	Redis:     "6379",
	Mongo:     "27017",
	SqlServer: "1433",
	Kafka:     "9092",

	KafkaSecond: "29092",
	EventBroker: "3000",
	Nginx:       "80",
	Zookeeper:   "2181",
}

var outPort = PortDto{
	Mysql:     "3306",
	Redis:     "6379",
	Mongo:     "27017",
	SqlServer: "1433",
	Kafka:     "9092",

	KafkaSecond: "29092",
	EventBroker: "3000",
	Nginx:       "3001",
	Zookeeper:   "2181",
}

type PortDto struct {
	Mysql       string
	Redis       string
	Mongo       string
	SqlServer   string
	Kafka       string
	KafkaSecond string

	EventBroker string
	Nginx       string
	Zookeeper   string
}
type ConfigDto struct {
	Scope   string
	Ip      string
	Port    PortDto
	Project *ProjectDto
}
type ProjectDto struct {
	IsMulti        bool     //a git contains multiple microservices
	ServiceName    string   //eg. ipay-api
	GitShortPath   string   //eg. ipay/ipay-api
	Envs           []string // from jenkins
	IsProjectKafka bool

	Ports            []string
	Databases        []string //mysql,redis,mongo,sqlserver
	StreamNames      []string
	ParentFolderName string
	Registry         string

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
		if err := (Relation{}).FetchsqlTofile(c.Project); err != nil {
			fmt.Println(err)
			return
		}
	}

	//2. generate docker-compose
	if shouldUpdateCompose(c.Scope) {
		viper := viper.New()
		compose := Compose{}
		if shouldStartKakfa(c.Project) {
			compose.setComposeKafkaEland(viper, c.Port.Kafka, c.Port.KafkaSecond, c.Port.Zookeeper, c.Ip)
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

		if err = writeToCompose(viper); err != nil {
			fmt.Println(err)
			return
		}
	}

	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	//3. run docker-compose
	if shouldRestartData(c.Scope) {
		//delete volume
		if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose downed!")
	}
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "pull"); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	fmt.Println("==> compose pulled!")

	if shouldRestartApp(c.Scope) {
		if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "build"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose builded!")
	}
	project := *(c.Project)
	go func(p ProjectDto, portD PortDto, composePath string) {
		if err = checkAll(p, portD, composePath); err != nil {
			fmt.Println(err)
		}
		fmt.Println("check is ok.")
		if _, err = CmdRealtime("docker-compose", "-f", composePath, "up", "-d"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		if _, err = CmdRealtime("docker", "ps", "-a"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose up!")
	}(project, c.Port, dockercompose)
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
	time.Sleep(100 * time.Hour)
}

func writeToCompose(viper *viper.Viper) (err error) {
	ymlStr, err := yamlStringSettings(viper)
	if err != nil {
		err = fmt.Errorf("write to %v error:%v\n", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		return
	}

	ymlStr = UpperKafkaEnvEland(ymlStr)

	if writeFile(TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", ymlStr); err != nil {
		err = fmt.Errorf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		return
	}
	return
}

func UpperKafkaEnvEland(ymlStr string) string {
	ymlStr = strings.Replace(ymlStr, "kafka_broker_id", "KAFKA_BROKER_ID", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_host_name", "KAFKA_ADVERTISED_HOST_NAME", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_port", "KAFKA_ADVERTISED_PORT", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connection_timeout_ms", "KAFKA_ZOOKEEPER_CONNECTION_TIMEOUT_MS", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_delete_topic_enable", "KAFKA_DELETE_TOPIC_ENABLE", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_dirs", "KAFKA_LOG_DIRS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_retention_hours", "KAFKA_LOG_RETENTION_HOURS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_cleanup_policy", "KAFKA_LOG_CLEANUP_POLICY", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_cleaner_enable", "KAFKA_LOG_CLEANER_ENABLE", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_jvm_performance_opts", "KAFKA_JVM_PERFORMANCE_OPTS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_heap_opts", "KAFKA_HEAP_OPTS", -1)
	ymlStr = strings.Replace(ymlStr, "jmx_port", "JMX_PORT", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_jmx_opts", "KAFKA_JMX_OPTS", -1)

	ymlStr = strings.Replace(ymlStr, "zoo_my_id", "ZOO_MY_ID", -1)
	ymlStr = strings.Replace(ymlStr, "zoo_servers", "ZOO_SERVERS", -1)

	ymlStr = UpperKafkaEnv(ymlStr)

	return ymlStr
}

func UpperKafkaEnv(ymlStr string) string {
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_listeners", "KAFKA_ADVERTISED_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_inter_broker_listener_name", "KAFKA_INTER_BROKER_LISTENER_NAME", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listener_security_protocol_map", "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listeners", "KAFKA_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_advertised_port", "KAFKA_ADVERTISED_PORT", -1)
	return ymlStr
}

func checkAll(project ProjectDto, port PortDto, dockercompose string) (err error) {

	if shouldStartMysql(&project) {
		if err = checkMysql(dockercompose, port.Mysql); err != nil {
			return
		}
	}
	if shouldStartKakfa(&project) {
		if err = checkKafka(dockercompose, port.Kafka); err != nil {
			return
		}
	}

	return
}

func checkMysql(dockercompose, port string) (err error) {

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "--detach", "mysql"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:1234@tcp(127.0.0.1:%v)/mysql?charset=utf8", port))
	if err != nil {
		fmt.Println("mysql", err)
		return
	}
	//remove mysql log
	buffer := bytes.NewBuffer(make([]byte, 0, 64))
	logger := log.New(buffer, "prefix: ", 0)
	mysql.SetLogger(logger)

	fmt.Println("begin ping db")
	for index := 0; index < 300; index++ {
		err = db.Ping()
		if err != nil {
			//fmt.Println("error ping db", err)
			time.Sleep(1 * time.Second)
			continue
		}
		err = nil
		break
	}
	if err != nil {
		fmt.Println("error ping db")
		return
	}
	fmt.Println("finish ping db")
	return
}

func checkKafka(dockercompose, port string) (err error) {

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "--detach", "zookeeper"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "--detach", "kafka"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	fmt.Println("begin ping kafka,localhost:" + port)
	for index := 0; index < 300; index++ {
		if err = dailKafka(port); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		err = nil
		break
	}
	if err != nil {
		fmt.Println("error ping kafka")
		return
	}
	fmt.Println("finish ping kafka")
	return
}

func dailKafka(port string) (err error) {
	_, err = kafkautil.DialLeader(context.Background(), "tcp", "localhost:"+port, "ping", 0)
	if err != nil {
		return
	}
	return
}
