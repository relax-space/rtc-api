package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
}

func (d Config) LoadEnv(serviceName string) (c *FullDto, err error) {
	if err = d.confirm(serviceName); err != nil {
		return
	}
	c = &FullDto{}
	if err = d.loadEnv(c); err != nil {
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

func (Config) readYmlRemote(serviceName string, c *FullDto) (err error) {
	//1.load base info from gitlab
	if c.Project, err = (Relation{}).FetchRelation(serviceName); err != nil {
		return
	}
	return
}

func (Config) currentScope(updated *string) (updatedStr string, err error) {
	if StringPointCheck(updated) {
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
		err = fmt.Errorf("Parameters(%v) are not supported, only support %v", updated, LOCAL.List())
		return
	}
	fmt.Printf("current:%v \n", updatedStr)
	return
}

func (d Config) readYml(serviceName string, c *FullDto) (err error) {
	if scope == LOCAL.String() {
		if err = (File{}).ReadViper("", c); err != nil {
			return
		}
		if len(serviceName) != 0 {
			c.Project.ServiceName = serviceName
		}
	}

	if err = d.readYmlRemote(serviceName, c); err != nil {
		return
	}
	return
}

func (d Config) loadEnv(c *FullDto) (err error) {

	updatedStr, err := Config{}.currentScope(envDto.Updated)
	if err != nil {
		err = fmt.Errorf("read env updated error:%v", err)
		return
	}
	scope = updatedStr

	if StringPointCheck(envDto.ImageEnv) {
		app_env = *envDto.ImageEnv
	} else {
		app_env = "qa"
	}

	c.Ip, err = getIp(envDto.Ip)
	if err != nil {
		return
	}
	if c.Project == nil {
		c.Project = &ProjectDto{}
	}
	c.Port.Mysql = *envDto.MysqlPort
	c.Port.Redis = *envDto.RedisPort
	c.Port.Mongo = *envDto.MongoPort
	c.Port.SqlServer = *envDto.SqlServerPort
	c.Port.Kafka = *envDto.KafkaPort

	c.Port.KafkaSecond = *envDto.KafkaSecondPort
	c.Port.Zookeeper = *envDto.ZookeeperPort
	c.Port.EventBroker = *envDto.EventBrokerPort
	c.Port.Nginx = *envDto.NginxPort

	return
}

func (d Config) confirm(serviceName string) (err error) {
	localServiceName, err := d.getServiceNameLocal()
	if err != nil {
		return
	}
	if len(localServiceName) == 0 {
		return
	}
	if len(serviceName) == 0 {
		return
	}
	if serviceName != localServiceName {
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
