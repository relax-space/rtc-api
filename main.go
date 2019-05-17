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
	Ip          string
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
	ServiceName: kingpin.Arg("name", "name from mingbai api.").String(),
	Updated:     kingpin.Flag("updated", "data from [local remote].").Short('u').Default("remote").String(),
	List:        kingpin.Flag("list", "Show all names from mingbai api.").Short('l').Bool(),
	Ip:          kingpin.Flag("ip", "IP address to connect internet.").IP().String(),
	ImageEnv:    kingpin.Flag("image-env", "image env [qa prd].").Default("qa").String(),

	MysqlPort:     kingpin.Flag("mysql-port", "set port mysql.").Default(outPort.Mysql).String(),
	RedisPort:     kingpin.Flag("redis-port", "set port redis.").Default(outPort.Redis).String(),
	MongoPort:     kingpin.Flag("mongo-port", "set port mongo.").Default(outPort.Mongo).String(),
	SqlServerPort: kingpin.Flag("sqlserver-port", "set port sqlserver.").Default(outPort.SqlServer).String(),
	KafkaPort:     kingpin.Flag("kafka-port", "set port kafka.").Default(outPort.Kafka).String(),

	KafkaSecondPort: kingpin.Flag("kafka-second-port", "set port kafka-second.").Default(outPort.KafkaSecond).String(),
	EventBrokerPort: kingpin.Flag("event-broker-port", "set port event-broker.").Default(outPort.EventBroker).String(),
	NginxPort:       kingpin.Flag("nginx-port", "set port nginx.").Default(outPort.Nginx).String(),
	ZookeeperPort:   kingpin.Flag("zookeeper-port", "set port zookeeper.").Default(outPort.Zookeeper).String(),
}

var Version string

func main() {
	if ok := Init(); ok == false {
		return
	}
	if err := (Config{}).CheckHost(envDto.Ip); err != nil {
		fmt.Println(err)
		os.Exit(1)
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

func Init() bool {
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
