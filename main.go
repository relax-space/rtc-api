package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-xorm/core"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"

	kafkautil "github.com/segmentio/kafka-go"

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
	P2SHOPHOST           = "https://gateway.p2shop.com.cn"
	QAREGISTRY           = "registry.p2shop.com.cn"
	WAITIMAGE            = "waisbrot/wait" //xiaoxinmiao/wait:0.0.2
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
	Mysql:     "3306",
	Redis:     "6379",
	Mongo:     "27017",
	SqlServer: "1433",
	Kafka:     "29092",

	EventBroker: "3000",
	Nginx:       "3001",
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

		if err = writeToCompose(viper); err != nil {
			fmt.Println(err)
			return
		}
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
	project := *(c.Project)
	go func(p ProjectDto, composePath string) {
		if err = checkAll(p, composePath); err != nil {
			fmt.Println(err)
		}
		fmt.Println("check is ok.")
		if _, err = Cmd("docker-compose", "-f", composePath, "up", "-d"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		time.Sleep(10 * time.Second)
		fmt.Println("==> compose may have been up!")
	}(project, dockercompose)
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

	ymlStr = strings.Replace(ymlStr, "kafka_advertised_listeners", "KAFKA_ADVERTISED_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_inter_broker_listener_name", "KAFKA_INTER_BROKER_LISTENER_NAME", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listener_security_protocol_map", "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listeners", "KAFKA_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)

	if writeFile(TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", ymlStr); err != nil {
		err = fmt.Errorf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		return
	}
	return
}

func checkAll(project ProjectDto, dockercompose string) (err error) {

	if shouldStartMysql(&project) {
		if err = checkMysql(dockercompose); err != nil {
			return
		}
	}
	if shouldStartKakfa(&project) {
		if err = checkKafka(dockercompose); err != nil {
			return
		}
	}

	return
}

func checkMysql(dockercompose string) (err error) {

	if _, err = Cmd("docker-compose", "-f", dockercompose, "up", "--detach", "mysql"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	db, err := xorm.NewEngine("mysql", fmt.Sprintf("root:1234@tcp(127.0.0.1:%v)/mysql?charset=utf8", outPort.Mysql))
	if err != nil {
		fmt.Println("mysql", err)
		return
	}

	db.SetLogLevel(core.LOG_OFF)

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

func checkKafka(dockercompose string) (err error) {

	if _, err = Cmd("docker-compose", "-f", dockercompose, "up", "--detach", "zookeeper"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	if _, err = Cmd("docker-compose", "-f", dockercompose, "up", "--detach", "kafka"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	fmt.Println("begin ping kafka")
	for index := 0; index < 300; index++ {
		_, err = kafkautil.DialLeader(context.Background(), "tcp", "localhost:"+outPort.Kafka, "ping", 0)
		if err != nil {
			time.Sleep(1 * time.Second)
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
