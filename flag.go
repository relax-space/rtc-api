package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
)

var Version string

type Flag struct {
	Updated     *string
	ImageEnv    *string
	Log         *bool
	IgnoreLogin *bool
	IgnorePull  *bool

	MysqlPort     *string
	RedisPort     *string
	MongoPort     *string
	SqlServerPort *string
	KafkaPort     *string

	KafkaSecondPort *string
	EventBrokerPort *string
	NginxPort       *string
	ZookeeperPort   *string
}

func (Flag) Init() (serviceName *string, flag *Flag) {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Author("qa group")
	kingpin.CommandLine.Help = "A tool that runs microservices and its dependencies.For detail flags of each command run `help [<command>...]`."
	kingpin.CommandLine.HelpFlag.Short('h')
	if len(Version) == 0 {
		Version = "v1.0"
	}
	kingpin.CommandLine.Version(Version).VersionFlag.Short('v')
	configureLsCommand(kingpin.CommandLine)
	serviceName, flag = configureRunCommand(kingpin.CommandLine)
	kingpin.Parse()
	return
}

func showList(q string) {
	list, err := Relation{}.FetchAllNames()
	if err != nil {
		log.Println(err)
		return
	}
	newList := make([]string, 0)
	if len(q) != 0 {
		for _, v := range list {
			vlow := strings.ToLower(v)
			if strings.Contains(vlow, strings.ToLower(q)) {
				newList = append(newList, v)
			}
		}
	} else {
		newList = list
	}
	if len(newList) == 0 {
		log.Println("no data has found.")
		return
	}
	for _, v := range newList {
		log.Println(v)
	}
}

func configureLsCommand(app *kingpin.Application) {
	var q string
	ls := kingpin.Command("ls", "List service names from remote.").Action(func(c *kingpin.ParseContext) error {
		showList(q)
		return nil
	})
	ls.Arg("q", "Fuzzy query service name by `q`").StringVar(&q)
}

func configureRunCommand(app *kingpin.Application) (serviceName *string, flag *Flag) {
	run := kingpin.Command("run", "Run a service and its dependencies.")
	pName := filepath.Base(os.Args[0])
	desc := fmt.Sprintf("The name of the service, you can get it by `%v ls -h`.", pName)
	serviceName = run.Arg("service-name", desc).Required().String()
	flag = &Flag{
		Updated: run.Flag("updated", `
	1.Optional [remote, local].
	2.The program will get the following information from the remote: service information,basic test data and docker image.
	3.The default is remote,if you don't want to get data from remote, please use local.`).Short('u').Default("remote").String(),
		ImageEnv: run.Flag("image-env", `
	1.Optional [staging, qa , production].
	2.The program will download the latest image from Jenkins.
	3.The default is qa, you can choose other option`).Default("qa").String(),

		Log:         run.Flag("log", "You can see log for debug.").Bool(),
		IgnoreLogin: run.Flag("ignore-login", "You can ignore login step.").Bool(),
		IgnorePull:  run.Flag("ignore-pull", "You can ignore pull images step.").Bool(),

		MysqlPort:     run.Flag("mysql-port", "You can change default mysql port.").Default(outPort.Mysql).String(),
		RedisPort:     run.Flag("redis-port", "You can change default redis port.").Default(outPort.Redis).String(),
		MongoPort:     run.Flag("mongo-port", "You can change default mongo port.").Default(outPort.Mongo).String(),
		SqlServerPort: run.Flag("sqlserver-port", "You can change default sqlserver port.").Default(outPort.SqlServer).String(),
		KafkaPort:     run.Flag("kafka-port", "You can change default kafka port.").Default(outPort.Kafka).String(),

		KafkaSecondPort: run.Flag("kafka-second-port", "This parameter is reserved.").Default(outPort.KafkaSecond).String(),
		EventBrokerPort: run.Flag("event-broker-port", "You can change default event-broker port.").Default(outPort.EventBroker).String(),
		NginxPort:       run.Flag("nginx-port", "You can change default nginx port.").Default(outPort.Nginx).String(),
		ZookeeperPort:   run.Flag("zookeeper-port", "You can change default zookeeper port.").Default(outPort.Zookeeper).String(),
	}
	return
}
