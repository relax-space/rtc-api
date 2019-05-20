package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
)

var envDto = &struct {
	Updated  *string
	ImageEnv *string

	MysqlPort     *string
	RedisPort     *string
	MongoPort     *string
	SqlServerPort *string
	KafkaPort     *string

	KafkaSecondPort *string
	EventBrokerPort *string
	NginxPort       *string
	ZookeeperPort   *string
}{
	Updated: kingpin.Flag("updated", `
	1.Optional [remote, local].
	2.The program will get the following information from the remote: project information,basic test data and docker image.
	3.The default is remote,if you don't want to get data from remote, please use local.`).Short('u').Default("remote").String(),
	ImageEnv: kingpin.Flag("image-env", `
	1.Optional [staging, qa , production].
	2.The program will download the latest image from Jenkins.
	3.The default is qa, you can choose other option`).Default("qa").String(),

	MysqlPort:     kingpin.Flag("mysql-port", "You can change default mysql port.").Default(outPort.Mysql).String(),
	RedisPort:     kingpin.Flag("redis-port", "You can change default redis port.").Default(outPort.Redis).String(),
	MongoPort:     kingpin.Flag("mongo-port", "You can change default mongo port.").Default(outPort.Mongo).String(),
	SqlServerPort: kingpin.Flag("sqlserver-port", "You can change default sqlserver port.").Default(outPort.SqlServer).String(),
	KafkaPort:     kingpin.Flag("kafka-port", "You can change default kafka port.").Default(outPort.Kafka).String(),

	KafkaSecondPort: kingpin.Flag("kafka-second-port", "This parameter is reserved.").Default(outPort.KafkaSecond).String(),
	EventBrokerPort: kingpin.Flag("event-broker-port", "You can change default event-broker port.").Default(outPort.EventBroker).String(),
	NginxPort:       kingpin.Flag("nginx-port", "You can change default nginx port.").Default(outPort.Nginx).String(),
	ZookeeperPort:   kingpin.Flag("zookeeper-port", "You can change default zookeeper port.").Default(outPort.Zookeeper).String(),
}

var (
	ls    = kingpin.Command("ls", "List project names from remote.")
	lsArg = ls.Arg("service-name-like", "Fuzzy query by `service-name-like`").String()

	run    = kingpin.Command("run", "Run a service and its dependencies.")
	runArg = run.Arg("service-name", "The name of the project, you can get it by `run-test ls -h`.").String()
)

type Flag struct {
}

func (Flag) Init() bool {
	exeName := filepath.Base(os.Args[0])
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Author("qa group")
	kingpin.CommandLine.Help = "A tool that runs microservices and its dependencies."
	kingpin.CommandLine.HelpFlag.Short('h')
	if len(Version) == 0 {
		Version = "v1.0"
	}
	kingpin.CommandLine.Version(Version).VersionFlag.Short('v')
	switch kingpin.Parse() {
	case "ls":
		showList()
		return false
	case "run":
		if StringPointCheck(runArg) == false {
			fmt.Printf("%v: error: required argument 'service-name' not provided, try `%v ls -h`", exeName, exeName)
			return false
		}
		return true
	}
	return true
}

func showList() {
	list, err := Relation{}.FetchAll()
	if err != nil {
		fmt.Println(err)
		return
	}
	newList := make([]string, 0)
	if StringPointCheck(lsArg) {
		for _, v := range list {
			vlow := strings.ToLower(v)
			if strings.Contains(vlow, strings.ToLower(*lsArg)) {
				newList = append(newList, v)
			}
		}
	} else {
		newList = list
	}
	if len(newList) == 0 {
		fmt.Println("no data has found.")
		return
	}
	for _, v := range newList {
		fmt.Println(v)
	}
}
