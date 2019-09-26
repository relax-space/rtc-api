package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ElandGroup/joblog"
	"github.com/alecthomas/kingpin"
)

var Version string

type Flag struct {
	LocalSql *bool
	Env      *string
	ImageEnv *string
	Debug    *bool
	NoLogin  *bool
	NoPull   *bool

	NoLog          *bool
	RegistryCommon *string
	HostIp         *string

	MysqlPort     *string
	RedisPort     *string
	MongoPort     *string
	SqlServerPort *string
	KafkaPort     *string

	KafkaSecondPort *string
	//EventBrokerPort *string
	NginxPort     *string
	ZookeeperPort *string
}

func (d Flag) Init() (isContinue bool, serviceName *string, flag *Flag) {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Author("qa group")
	kingpin.CommandLine.Help = "A tool that runs microservices and its dependencies.For detail flags of each command run `help [<command>...]`."
	if len(Version) == 0 {
		Version = "v1.0"
	}
	kingpin.CommandLine.Version(Version)
	d.configureLsCommand(kingpin.CommandLine)
	d.configureDownCommand(kingpin.CommandLine)
	serviceName, flag = d.configureRunCommand(kingpin.CommandLine)
	switch kingpin.Parse() {
	case "ls", "down":
		isContinue = false
	default:
		isContinue = true
		SetEnv(*flag.Env)
		if StringPointCheck(serviceName) == false {
			panic("service name is required.")
		}
		if BoolPointCheck(flag.Debug) {
			log.SetFlags(log.Lshortfile | log.LstdFlags)
		} else {
			log.SetFlags(0)
		}
		flag.HostIp = d.getIp(flag.HostIp)
		if BoolPointCheck(flag.NoLog) == false {
			log.Println("log init ...")
			if err := d.initJobLog(serviceName, flag); err != nil {
				log.Println(err)
				panic(err)
			}
		}
	}
	return
}

func (d Flag) showList(q string) error {
	names, err := Project{}.GetServiceNames(q)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return errors.New("no data has found.")
	}
	for _, v := range names {
		fmt.Println(v)
	}
	return nil
}

func (d Flag) configureLsCommand(app *kingpin.Application) {
	var q string
	var env *string
	ls := kingpin.Command("ls", "List service names from remote.").Action(func(c *kingpin.ParseContext) error {
		SetEnv(*env)
		err := d.showList(q)
		if err != nil {
			panic(err)
		}
		return nil
	})
	ls.Arg("q", "Fuzzy query service name by `q`").StringVar(&q)
	env = ls.Flag("env", `
	1.Optional [staging, qa , production].
	2.rtc runtime environment variable.
	3.The default is qa, you can choose other option`).Default("qa").String()
}

func (d Flag) configureDownCommand(app *kingpin.Application) {
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

func (d Flag) configureRunCommand(app *kingpin.Application) (serviceName *string, flag *Flag) {
	outPort := PortDto{
		Mysql:     "3308",
		Redis:     "6381",
		Mongo:     "27019",
		SqlServer: "1435",
		Kafka:     "9092",

		KafkaSecond: "29092",
		//EventBroker: "3002",
		Nginx: "3001",
		//Zookeeper:   "2181",
	}

	run := kingpin.Command("run", "Run a service and its dependencies.")
	pName := filepath.Base(os.Args[0])
	desc := fmt.Sprintf("The name of the service, you can get it by `./%v ls`.", pName)
	serviceName = run.Arg("service-name", desc).Required().String()
	flag = &Flag{
		LocalSql: run.Flag("local-sql", `Load data from a local file.`).Bool(),
		ImageEnv: run.Flag("image-env", `
	1.Optional [staging, qa , production].
	2.microservice docker image runtime environment variable.
	3.The default is qa, you can choose other option`).Default("qa").String(),
		Env: run.Flag("env", `
	1.Optional [staging, qa , production].
	2.rtc runtime environment variable.
	3.The default is qa, you can choose other option`).Default("qa").String(),

		Debug:   run.Flag("debug", "You can see log for debug.").Bool(),
		NoLogin: run.Flag("no-login", "You can ignore login step.").Bool(),
		NoPull:  run.Flag("no-pull", "You can ignore pull images step.").Bool(),
		NoLog:   run.Flag("no-log", "You can disable uploading logs.").Bool(),
		RegistryCommon: run.Flag("registry-common", `
	1.You can set private registry for common image,like: mysql,ngnix,kafka.
	2.default: registry.p2shop.com.cn.`).String(),
		HostIp: run.Flag("host-ip", `
	1.ip(default): auto get ip.
	2.You can specify your host ip.`).String(),

		MysqlPort:     run.Flag("mysql-port", "You can change default mysql port.").Default(outPort.Mysql).String(),
		RedisPort:     run.Flag("redis-port", "You can change default redis port.").Default(outPort.Redis).String(),
		MongoPort:     run.Flag("mongo-port", "You can change default mongo port.").Default(outPort.Mongo).String(),
		SqlServerPort: run.Flag("sqlserver-port", "You can change default sqlserver port.").Default(outPort.SqlServer).String(),
		KafkaPort:     run.Flag("kafka-port", "You can change default kafka port.").Default(outPort.Kafka).String(),

		KafkaSecondPort: run.Flag("kafka-second-port", "This parameter is reserved.").Default(outPort.KafkaSecond).String(),
		//EventBrokerPort: run.Flag("event-broker-port", "You can change default event-broker port.").Default(outPort.EventBroker).String(),
		NginxPort: run.Flag("nginx-port", "You can change default nginx port.").Default(outPort.Nginx).String(),
		//ZookeeperPort:   run.Flag("zookeeper-port", "You can change default zookeeper port.").Default(outPort.Zookeeper).String(),
	}
	return
}

func (d Flag) getIp(ipFlag *string) *string {
	if StringPointCheck(ipFlag) {
		return ipFlag
	}
	ip, err := currentIp()
	if err != nil {
		log.Println("WARNING: fetch ip failure, has set ip to 127.0.0.1")
		return func(ip string) *string { return &ip }(ip)
	}
	return &ip
}

func (d Flag) initJobLog(serviceName *string, flag *Flag) error {

	jobLog = joblog.New(JobLogUrl+"/batchjob-api/v1/jobs", "rtc-api", map[string]interface{}{"service name:": serviceName, "ip": *flag.HostIp})
	if jobLog.Err != nil {
		return jobLog.Err
	}
	jobLog.Info(flag)
	return nil
}
