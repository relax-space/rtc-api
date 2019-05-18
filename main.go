package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

var envDto = &struct {
	ServiceName *string
	Updated     *string
	List        *bool
	ImageEnv    *string

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
	ServiceName: kingpin.Arg("name", "The name of the project, you can get it by --list.").String(),
	Updated: kingpin.Flag("updated", `
	1.Optional [remote, local].
	2.The program will get the following information from the remote: project information,basic test data and docker image.
	3.The default is remote,if you don't want to get data from remote, please use local.`).Short('u').Default("remote").String(),
	List: kingpin.Flag("list", "Query project name from remote.").Short('l').Bool(),
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

var Version string

func main() {
	if ok := InitFlag(); ok == false {
		return
	}

	c, err := Config{}.LoadEnv(*envDto.ServiceName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = writeLocal(c); err != nil {
		fmt.Println(err)
		return
	}

	if err = (Compose{}).Exec(c); err != nil {
		fmt.Println(err)
		return
	}
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

func writeLocal(c *FullDto) (err error) {
	if scope == LOCAL.String() {
		return
	}
	if err = composeWriteYml(c); err != nil {
		return
	}
	if err = (Nginx{}).WriteConfig(c.Project, c.Port.EventBroker); err != nil {
		return
	}
	if err = (ProjectInfo{}).WriteSql(c.Project); err != nil {
		return
	}

	if err = (Config{}).WriteYml(c); err != nil {
		return
	}
	return
}

func composeWriteYml(c *FullDto) (err error) {
	viper := viper.New()
	p := ProjectInfo{}
	d := Compose{}
	if p.ShouldKafka(c.Project) {
		d.setComposeKafkaEland(viper, c.Port.Kafka, c.Port.KafkaSecond, c.Port.Zookeeper, c.Ip)
	}
	if p.ShouldDb(c.Project, MYSQL) {
		d.setComposeMysql(viper, c.Port.Mysql)
	}
	if p.ShouldDb(c.Project, REDIS) {
		d.setComposeRedis(viper, c.Port.Redis)
	}

	if p.ShouldEventBroker(c.Project) {
		streamNames := p.StreamList(c.Project)
		EventBroker{}.SetEventBroker(viper, c.Port.EventBroker, streamNames)
	}
	d.setComposeApp(viper, c.Project)
	d.setComposeNginx(viper, c.Project.ServiceName, c.Port.Nginx)
	d.WriteYml(viper)
	return
}

func InitFlag() bool {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Author("qa group")
	kingpin.CommandLine.Help = "A tool that runs microservices and its dependencies."
	kingpin.CommandLine.HelpFlag.Short('h')
	if len(Version) == 0 {
		Version = "v1.0"
	}
	kingpin.CommandLine.Version(Version).VersionFlag.Short('v')
	kingpin.Parse()
	if BoolPointCheck(envDto.List) {
		list, err := Relation{}.FetchAll()
		if err != nil {
			fmt.Println(err)
			return false
		}
		for _, v := range list {
			fmt.Println(v)
		}
		return false
	}
	if StringPointCheck(envDto.ServiceName) == false {
		fmt.Println("error: required argument 'name' not provided, try --help")
		return false
	}

	return true
}
