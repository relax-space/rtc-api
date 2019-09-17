package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
)

var Version string

type Flag struct {
	Updated  *string
	ImageEnv *string
	Debug    *bool
	NoLogin  *bool
	NoPull   *bool

	RelationSource *bool
	ComboResource  *string
	NoLog          *bool
	RegistryCommon *string
	HostIp         *string

	MysqlPort     *string
	RedisPort     *string
	MongoPort     *string
	SqlServerPort *string
	KafkaPort     *string

	KafkaSecondPort *string
	EventBrokerPort *string
	NginxPort       *string
	ZookeeperPort   *string

	UrlGitlab *string
	PrivateTokenGitlab *string
}

func (Flag) Init() (isContinue bool, serviceName *string, flagParam *Flag) {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Author("qa group")
	kingpin.CommandLine.Help = "A tool that runs microservices and its dependencies.For detail flags of each command run `help [<command>...]`."
	if len(Version) == 0 {
		Version = "v1.0"
	}
	kingpin.CommandLine.Version(Version)
	configureLsCommand(kingpin.CommandLine)
	configureDownCommand(kingpin.CommandLine)
	serviceName, flagParam = configureRunCommand(kingpin.CommandLine)
	switch kingpin.Parse() {
	case "ls", "down":
		isContinue = false
	default:
		isContinue = true
	}
	return
}

func showList(q string, r *bool) (err error) {
	list, err := Relation{}.FetchAllNames(r)
	if err != nil {
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
		err = errors.New("no data has found.")
		return
	}
	for _, v := range newList {
		fmt.Println(v)
	}
	return
}

func configureLsCommand(app *kingpin.Application) {
	var q string
	var r *bool
	ls := kingpin.Command("ls", "List service names from remote.").Action(func(c *kingpin.ParseContext) error {
		err := showList(q, r)
		if err != nil {
			panic(err)
		}
		return nil
	})
	ls.Arg("q", "Fuzzy query service name by `q`").StringVar(&q)
	r = ls.Flag("relation-source", `
	1.false: default,fetch relation from https://gitlab.p2shop.cn:8443/data/rtc-data.
	2.true:fetch relation from mingbai-api`).Short('r').Bool()
}

func configureDownCommand(app *kingpin.Application) {
	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	var rmi *string
	var v *bool
	var remove *bool
	var t *int
	down := kingpin.Command("down", `
	Stops containers and removes containers, networks, volumes, and images
	created by 'up'.

	By default, the only things removed are:

	- Containers for services defined in the Compose file
	- Networks defined in the 'networks' section of the Compose file
	- The default network, if one is used

	Networks and volumes defined as 'external' are never removed.`).
		Action(func(c *kingpin.ParseContext) error {
			param := make([]string, 0)
			param = append(param, "-f", dockercompose, "down")
			if StringPointCheck(rmi) {
				param = append(param, "--rmi", *rmi)
			}
			if BoolPointCheck(v) {
				param = append(param, "-v")
			}
			if BoolPointCheck(remove) {
				param = append(param, "--remove-orphans")
			}
			if IntPointCheck(t) {
				param = append(param, "--timeout", fmt.Sprint(*t))
			}
			if _, err := CmdRealtime("docker-compose", param...); err != nil {
				panic(err)
			}
			return nil
		})
	rmi = down.Flag("rmi", `
    Remove images. Type must be one of:
        'all': Remove all images used by any service.
        'local': Remove only images that don't have a
    	custom tag set by the 'image' field.`).String()
	v = down.Flag("volumes", `
    Remove named volumes declared in the 'volumes'
    section of the Compose file and anonymous volumes
    attached to containers`).Short('v').Bool()
	remove = down.Flag("remove-orphans", `
    Remove containers for services not defined in the
    Compose file`).Bool()
	t = down.Flag("timeout", `
    Specify a shutdown timeout in seconds.
    (default: 10)`).Short('t').Int()
}

func configureRunCommand(app *kingpin.Application) (serviceName *string, flag *Flag) {
	run := kingpin.Command("run", "Run a service and its dependencies.")
	pName := filepath.Base(os.Args[0])
	desc := fmt.Sprintf("The name of the service, you can get it by `./%v ls`.", pName)
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

		Debug:   run.Flag("debug", "You can see log for debug.").Bool(),
		NoLogin: run.Flag("no-login", "You can ignore login step.").Bool(),
		NoPull:  run.Flag("no-pull", "You can ignore pull images step.").Bool(),
		NoLog:   run.Flag("no-log", "You can disable uploading logs.").Bool(),
		RegistryCommon: run.Flag("registry-common", `
	1.You can set private registry.
	2.default: registry.p2shop.com.cn.`).String(),
		HostIp: run.Flag("host-ip", `
	1.ip(default): auto get ip.
	2.You can specify your host ip.`).String(),

		RelationSource: run.Flag("relation-source", `
	1.false: default,fetch relation from gitlab,like https://gitlab.p2shop.cn:8443/data/rtc-data.
	2.true:fetch relation from mingbai-api`).Short('r').Bool(),
		ComboResource: run.Flag("combo-resource", `
	1.Optional [p2shop, srx , srx-p2shop].
	2.p2shop git:https://gitlab.p2shop.cn:8443 jenkins:https://ci.p2shop.com.cn
	  srx git:https://gitlab.srxcloud.com jenkins:https://jenkins.srxcloud.com
	 srx-p2shop git:https://gitlab.srxcloud.com jenkins:https://ci.p2shop.com.cn
	3.The default is srx-p2shop, you can choose other option`).Default("srx-p2shop").Short('c').String(),

		MysqlPort:     run.Flag("mysql-port", "You can change default mysql port.").Default(outPort.Mysql).String(),
		RedisPort:     run.Flag("redis-port", "You can change default redis port.").Default(outPort.Redis).String(),
		MongoPort:     run.Flag("mongo-port", "You can change default mongo port.").Default(outPort.Mongo).String(),
		SqlServerPort: run.Flag("sqlserver-port", "You can change default sqlserver port.").Default(outPort.SqlServer).String(),
		KafkaPort:     run.Flag("kafka-port", "You can change default kafka port.").Default(outPort.Kafka).String(),

		KafkaSecondPort: run.Flag("kafka-second-port", "This parameter is reserved.").Default(outPort.KafkaSecond).String(),
		EventBrokerPort: run.Flag("event-broker-port", "You can change default event-broker port.").Default(outPort.EventBroker).String(),
		NginxPort:       run.Flag("nginx-port", "You can change default nginx port.").Default(outPort.Nginx).String(),
		ZookeeperPort:   run.Flag("zookeeper-port", "You can change default zookeeper port.").Default(outPort.Zookeeper).String(),
		UrlGitlab:   run.Flag("url-gitlab", "You can change default zookeeper port.").String(),
		PrivateTokenGitlab:   run.Flag("private-token-gitlab", "You can change default zookeeper port.").String(),
	}
	return
}
