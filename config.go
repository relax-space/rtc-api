package main

import (
	"fmt"
	"strings"

	"github.com/lextoumbourou/goodhosts"

	"github.com/spf13/viper"
)

type Config struct {
}

func (d Config) LoadEnv(serviceName string, flag *Flag) (c *FullDto, err error) {
	if err = d.confirm(serviceName, flag); err != nil {
		return
	}
	ip, err := currentIp()
	if err != nil {
		return
	}
	if err = (Config{}).CheckHost(ip); err != nil {
		return
	}
	c = &FullDto{
		Ip: ip,
	}
	if err = d.loadEnv(c, flag); err != nil {
		return
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

func (Config) CheckHost(ip string) (err error) {
	mapHost := map[string]string{
		"10.202.101.200": "registry.elandsystems.cn",
		ip:               "test-kafka",
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

func (d Config) loadEnv(c *FullDto, flag *Flag) (err error) {
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
	c.Port.Mysql = *flag.MysqlPort
	c.Port.Redis = *flag.RedisPort
	c.Port.Mongo = *flag.MongoPort
	c.Port.SqlServer = *flag.SqlServerPort
	c.Port.Kafka = *flag.KafkaPort

	c.Port.KafkaSecond = *flag.KafkaSecondPort
	c.Port.Zookeeper = *flag.ZookeeperPort
	c.Port.EventBroker = *flag.EventBrokerPort
	c.Port.Nginx = *flag.NginxPort

	return
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
