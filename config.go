package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv      *string
	ServiceName *string
	Updated     *string
	Ip          *string
	MysqlPort   *string

	RedisPort       *string
	MongoPort       *string
	SqlServerPort   *string
	KafkaPort       *string
	KafkaSecondPort *string

	EventBrokerPort *string
	NginxPort       *string
	ZookeeperPort   *string
}

func (d Config) LoadEnv() (c *FullDto, err error) {

	appEnv := flag.String("appEnv", os.Getenv("appEnv"), "appEnv")
	serviceName := flag.String("s", os.Getenv("s"), "serviceName from mingbai")
	updated := flag.String("u", os.Getenv("u"), "remote or local")
	ip := flag.String("ip", os.Getenv("ip"), "ip")
	mysqlPort := flag.String("mysqlPort", os.Getenv("mysqlPort"), "mysqlPort")

	redisPort := flag.String("redisPort", os.Getenv("redisPort"), "redisPort")
	mongoPort := flag.String("mongoPort", os.Getenv("mongoPort"), "mongoPort")
	sqlServerPort := flag.String("sqlServerPort", os.Getenv("sqlServerPort"), "sqlServerPort")
	kafkaPort := flag.String("kafkaPort", os.Getenv("kafkaPort"), "kafkaPort")
	kafkaSecondPort := flag.String("kafkaSecondPort", os.Getenv("kafkaSecondPort"), "kafkaSecondPort")

	eventBrokerPort := flag.String("eventBrokerPort", os.Getenv("eventBrokerPort"), "eventBrokerPort")
	nginxPort := flag.String("nginxPort", os.Getenv("nginxPort"), "nginxPort")
	zookeeperPort := flag.String("zookeeperPort", os.Getenv("zookeeperPort"), "zookeeperPort")

	flag.Parse()

	if err = d.confirm(serviceName); err != nil {
		return
	}

	c = &FullDto{}
	config := &Config{
		AppEnv:      appEnv,
		ServiceName: serviceName,
		Updated:     updated,
		Ip:          ip,
		MysqlPort:   mysqlPort,

		RedisPort:       redisPort,
		MongoPort:       mongoPort,
		SqlServerPort:   sqlServerPort,
		KafkaPort:       kafkaPort,
		KafkaSecondPort: kafkaSecondPort,

		EventBrokerPort: eventBrokerPort,
		NginxPort:       nginxPort,
		ZookeeperPort:   zookeeperPort,
	}
	if err = config.loadEnv(c); err != nil {
		return
	}

	err = d.readYml(serviceName, c)
	if err != nil {
		return
	}
	return
}

func (Config) WriteYml(c *FullDto) (err error) {

	fileName := TEMP_FILE + "/" + YMLNAMECONFIG + ".yml"
	vip := viper.New()
	vip.SetConfigName(YMLNAMECONFIG)
	vip.AddConfigPath(TEMP_FILE)
	vip.Set("ip", c.Ip)
	vip.Set("port", c.Port)
	vip.Set("project", c.Project)
	err = File{}.WriteViper(fileName, vip)
	if err != nil {
		err = fmt.Errorf("write to config.yml error:%v", err)
		return
	}
	return
}

// private method ===========

func (Config) readYmlRemote(serviceName *string, c *FullDto) (err error) {
	if serviceName == nil || len(*serviceName) == 0 {
		err = fmt.Errorf("read env error:%v", "serviceName is required.")
		return
	}
	//1.load base info from gitlab
	if c.Project, err = (Relation{}).FetchRelation(*serviceName); err != nil {
		return
	}
	return
}

func (Config) currentScope(updated *string) (updatedStr string, err error) {
	if updated == nil || len(*updated) == 0 {
		updatedStr = REMOTE.String()
		return
	}
	has, err := (File{}).IsExist(TEMP_FILE + "/" + YMLNAMECONFIG + ".yml")
	if err != nil {
		return
	}
	if has == false {
		updatedStr = REMOTE.String()
		return
	}
	for _, s := range LOCAL.List() {
		if strings.ToLower(*updated) == s {
			updatedStr = s
			break
		}
	}
	if len(updatedStr) == 0 {
		err = fmt.Errorf("Parameters(%v) are not supported, only support %v", *updated, LOCAL.List())
		return
	}
	fmt.Printf("current:%v \n", updatedStr)
	return
}

func (d Config) readYml(serviceName *string, c *FullDto) (err error) {
	if scope == LOCAL.String() {
		if err = (File{}).ReadViper("", c); err != nil {
			return
		}
		if serviceName != nil && len(*serviceName) != 0 {
			c.Project.ServiceName = *serviceName
		}
	}

	if err = d.readYmlRemote(serviceName, c); err != nil {
		return
	}
	return
}

func (d *Config) loadEnv(c *FullDto) (err error) {

	updatedStr, err := Config{}.currentScope(d.Updated)
	if err != nil {
		err = fmt.Errorf("read env updated error:%v", err)
		return
	}
	scope = updatedStr

	if d.AppEnv != nil && len(*d.AppEnv) != 0 {
		app_env = *d.AppEnv
	} else {
		app_env = "qa"
	}

	c.Ip, err = getIp(d.Ip)
	if err != nil {
		return
	}
	if c.Project == nil {
		c.Project = &ProjectDto{}
	}
	if d.MysqlPort == nil || len(*d.MysqlPort) == 0 {
		c.Port.Mysql = inPort.Mysql
	} else {
		c.Port.Mysql = *d.MysqlPort
	}
	if d.RedisPort == nil || len(*d.RedisPort) == 0 {
		c.Port.Redis = inPort.Redis
	} else {
		c.Port.Mysql = *d.RedisPort
	}
	if d.MongoPort == nil || len(*d.MongoPort) == 0 {
		c.Port.Mongo = inPort.Mongo
	} else {
		c.Port.Mongo = *d.MongoPort
	}
	if d.SqlServerPort == nil || len(*d.SqlServerPort) == 0 {
		c.Port.SqlServer = inPort.SqlServer
	} else {
		c.Port.SqlServer = *d.SqlServerPort
	}
	if d.KafkaPort == nil || len(*d.KafkaPort) == 0 {
		c.Port.Kafka = inPort.Kafka
	} else {
		c.Port.Kafka = *d.KafkaPort
	}
	if d.KafkaSecondPort == nil || len(*d.KafkaSecondPort) == 0 {
		c.Port.KafkaSecond = inPort.KafkaSecond
	} else {
		c.Port.KafkaSecond = *d.KafkaSecondPort
	}
	if d.ZookeeperPort == nil || len(*d.ZookeeperPort) == 0 {
		c.Port.Zookeeper = inPort.Zookeeper
	} else {
		c.Port.Zookeeper = *d.ZookeeperPort
	}

	if d.EventBrokerPort == nil || len(*d.EventBrokerPort) == 0 {
		c.Port.EventBroker = inPort.EventBroker
	} else {
		c.Port.EventBroker = *d.EventBrokerPort
	}
	// nginx default outPort:3001
	if d.NginxPort == nil || len(*d.NginxPort) == 0 {
		c.Port.Nginx = "3001"
	} else {
		c.Port.Nginx = *d.NginxPort
	}

	return
}

func (d Config) confirm(serviceName *string) (err error) {
	localServiceName, err := d.getServiceNameLocal()
	if err != nil {
		return
	}
	if len(localServiceName) == 0 {
		return
	}
	if serviceName == nil || len(*serviceName) == 0 {
		return
	}
	if *serviceName != localServiceName {
		warning := `WARNING! This will remove all files in temp [y/N]?`
		if err = scan(warning); err != nil {
			return
		}
		if err = (File{}).DeleteAll("./" + TEMP_FILE + "/"); err != nil {
			return
		}
	}
	return
}

func (Config) getServiceNameLocal() (serviceName string, err error) {
	has, err := File{}.IsExist(TEMP_FILE + "/" + YMLNAMECONFIG + ".yml")
	if err != nil {
		return
	}
	if has == false {
		return
	}
	c := &FullDto{}
	if err = (File{}).ReadViper("", c); err != nil {
		err = fmt.Errorf("read config error:%v", err)
		return
	}
	serviceName = c.Project.ServiceName
	return
}
