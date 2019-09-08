package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
}

func (d Config) LoadEnv(serviceName, ip string, flag *Flag) (c *FullDto, err error) {
	if err = d.confirm(serviceName, flag); err != nil {
		return
	}
	c = &FullDto{
		Ip: ip,
	}
	if err = d.LoadFlag(c, flag); err != nil {
		return
	}
	if scope != LOCAL.String() {
		if err = (File{}).DeleteAll("./" + TEMP_FILE + "/"); err != nil {
			return
		}
	}

	err = d.readYml(serviceName, c, flag)
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

func (Config) readYmlRemote(serviceName string, c *FullDto, flag *Flag) (err error) {
	//1.load base info from gitlab
	if c.Project, err = (Relation{}).FetchProject(serviceName, flag); err != nil {
		return
	}
	return
}

func (Config) currentScope(updated *string) (updatedStr string, err error) {
	if StringPointCheck(updated) == false {
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
	return
}

func (d Config) readYml(serviceName string, c *FullDto, flag *Flag) (err error) {
	if scope == LOCAL.String() {
		if err = (File{}).ReadViper("", c); err != nil {
			return
		}
		if len(serviceName) != 0 {
			c.Project.ServiceName = serviceName
		}
		return
	}

	if err = d.readYmlRemote(serviceName, c, flag); err != nil {
		return
	}
	return
}

func (d Config) LoadFlag(c *FullDto, flag *Flag) (err error) {
	updatedStr, err := Config{}.currentScope(flag.Updated)
	if err != nil {
		err = fmt.Errorf("read env updated error:%v", err)
		return
	}
	fmt.Printf("current:%v \n", updatedStr)

	scope = updatedStr

	if StringPointCheck(flag.ImageEnv) {
		app_env = *flag.ImageEnv
	} else {
		app_env = "qa"
	}

	if c.Project == nil {
		c.Project = &ProjectDto{}
	}
	c.Port = d.LoadFlagPort(flag)
	return
}

func (d Config) LoadFlagPort(flag *Flag) PortDto {
	port := PortDto{}
	port.Mysql = *flag.MysqlPort
	port.Redis = *flag.RedisPort
	port.Mongo = *flag.MongoPort
	port.SqlServer = *flag.SqlServerPort
	port.Kafka = *flag.KafkaPort

	port.KafkaSecond = *flag.KafkaSecondPort
	port.Zookeeper = *flag.ZookeeperPort
	port.EventBroker = *flag.EventBrokerPort
	port.Nginx = *flag.NginxPort
	return port
}
func (d Config) confirm(serviceName string, flag *Flag) (err error) {
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
		_, err = Relation{}.Fetch(serviceName, flag)
		if err != nil {
			return
		}
		if err = d.deleteTempFileWithTip(); err != nil {
			return
		}
	}
	return
}

func (Config) deleteTempFileWithTip() (err error) {
	warning := `WARNING! This will remove all files in temp [y/N]?`
	if err = scan(warning); err != nil {
		return
	}
	if err = (File{}).DeleteAll("./" + TEMP_FILE + "/"); err != nil {
		return
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
