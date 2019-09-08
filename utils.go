package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/lextoumbourou/goodhosts"

	"github.com/ElandGroup/joblog"
	"github.com/spf13/viper"
)

var (
	app_env = ""
	scope   = ""
)
var jobLog *joblog.JobLog

const (
	RtcPreGitUrl    = "https://gitlab.p2shop.cn:8443"
	RtcPrivateToken = "bY2kmqs8x8N3wfQxgw6s"

	YMLNAMECONFIG        = "config"
	YMLNAMEPROJEC        = "project"
	YMLNAMEDOCKERCOMPOSE = "docker-compose"
	CONFIGNAMENGINX      = "default"
	SUFSERVER            = "-server"
	PRETEST              = "test-"

	TEST_INFO        = "test_info"
	TEMP_FILE        = "temp"
	EventBroker_Name = "event-broker-kafka"
	jobLogUrl        = "https://gateway.p2shop.com.cn/batchjob-api/v1/jobs"
)

type IpPortDto struct {
	Ip   string
	Port string
}

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
	Mysql:     "3308",
	Redis:     "6381",
	Mongo:     "27019",
	SqlServer: "1435",
	Kafka:     "9092",

	KafkaSecond: "29092",
	EventBroker: "3002",
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
type FullDto struct {
	//Scope   string
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
	ExecPath       string
	Entrypoint     string

	Ports       []string
	Databases   map[string][]string //mysql,redis,mongo,sqlserver
	StreamNames []string
	Registry    string

	SubProjects []*ProjectDto
	DependsOn   []string
}

func CmdRealtime(name string, arg ...string) (result string, err error) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}
	return
}

func ContainString(chars []string, name string) bool {
	for _, c := range chars {
		if strings.ToUpper(c) == strings.ToUpper(name) {
			return true
		}
	}
	return false
}

func getStringViper(vip *viper.Viper) (ymlString string, err error) {
	c := vip.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		return
	}
	ymlString = string(bs)
	return
}

func currentIp() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = localAddr.IP.String()
	return
}

func scan(message string) (err error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message + ": ")
	b, _, err := reader.ReadLine()
	if err != nil {
		return
	}
	text := string(b)
	text = strings.ToLower(text)
	if ContainString(N.List(), text) == false {
		scan(message)
	}
	if text == N.String() {
		err = errors.New("user canceled! ")
		return
	}
	err = nil
	return
}

func BoolPointCheck(b *bool) (flag bool) {
	if b == nil || *b == false {
		return
	}
	flag = true
	return
}

func StringPointCheck(s *string) (flag bool) {
	if s == nil || len(*s) == 0 {
		return
	}
	flag = true
	return
}

func Unique(params []string) (list []string) {
	list = make([]string, 0)
	temp := make(map[string]string, 0)
	for _, p := range params {
		temp[p] = ""
	}
	for k := range temp {
		list = append(list, k)
	}
	return
}

func CurrentDatetime() string {
	return time.Now().Format("20060102150405")
}

func Error(err error) {
	if jobLog != nil {
		jobLog.Error(err)
	}
	panic(err)
}

func Info(message interface{}) {
	log.Println(message)
	if jobLog != nil {
		jobLog.Info(message)
	}
}

func CheckHost(ip string) (err error) {
	if runtime.GOOS != "windows" {
		return
	}
	mapHost := map[string]string{
		//"10.202.101.200": "registry.elandsystems.cn",
		ip: "test-kafka",
	}

	hosts, err := goodhosts.NewHosts()
	if err != nil {
		return
	}

	message := ""
	for k, v := range mapHost {
		if hosts.Has(k, v) == false {
			message += fmt.Sprintf("%v %v\n", k, v)
		}
	}
	if len(message) != 0 {
		err = fmt.Errorf("Please manually set the host file: \n%v", message)
		return
	}

	return
}
