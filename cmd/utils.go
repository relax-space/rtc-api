package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/lextoumbourou/goodhosts"

	"github.com/ElandGroup/joblog"
	"github.com/spf13/viper"
)

var jobLog *joblog.JobLog

const (
	YMLNAMEDOCKERCOMPOSE = "docker-compose"
	CONFIGNAMENGINX      = "default"
	SUFSERVER            = "-server"

	//EventBroker_Name = "event-broker-kafka"
	JobLogUrl      = "https://gateway.p2shop.com.cn"
	REGISTRYCOMMON = "registry.p2shop.com.cn"
)

var TEMP_FILE = "temp"

type PortDto struct {
	Mysql       string
	Redis       string
	Mongo       string
	SqlServer   string
	Kafka       string
	KafkaSecond string

	//EventBroker string
	Nginx     string
	Zookeeper string
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
func BoolPoint(b bool) *bool {
	return &b
}
func BoolPointCheck(b *bool) (flag bool) {
	if b == nil || *b == false {
		return
	}
	flag = true
	return
}

func IntPointCheck(b *int) (flag bool) {
	if b == nil || *b == 0 {
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

func CheckHost(ip, prefix string) (err error) {
	mapHost := map[string]string{
		//"10.202.101.200": "registry.elandsystems.cn",
		ip: prefix + "kafka",
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

func AddQuote(name string) string {
	return "\"" + name + "\""
}
