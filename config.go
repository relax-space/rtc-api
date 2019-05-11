package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func LoadEnv() (c *ConfigDto, err error) {

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

	//EventBroker
	flag.Parse()

	updatedStr, err := getScope(updated)
	if err != nil {
		err = fmt.Errorf("read env updated error:%v", err)
		return
	}

	c = &ConfigDto{}
	if err = loadEnv(c, updatedStr, ip, appEnv,
		mysqlPort, redisPort, mongoPort, sqlServerPort, kafkaPort, kafkaSecondPort,
		eventBrokerPort, nginxPort, zookeeperPort); err != nil {
		return
	}
	isLocalConfig := scopeSettings(updatedStr)

	if isLocalConfig {
		fmt.Printf("current:%v \n", LOCAL.String())
	} else {
		fmt.Printf("current:%v \n", REMOTE.String())
	}

	if isLocalConfig {
		if err = Read("", c); err != nil {
			err = fmt.Errorf("read config error:%v", err)
			return
		}
		if len(c.Project.GitShortPath) == 0 {
			fmt.Printf("no data from local temp/config.yml,please check param -s=%v", c.Project.ServiceName)
			return
		}
		return
	}

	if serviceName == nil || len(*serviceName) == 0 {
		err = fmt.Errorf("read env error:%v", "serviceName is required.")
		return
	}
	if err = deleteAllFile("./" + TEMP_FILE + "/"); err != nil {
		fmt.Println(err)
		return
	}
	//1.load base info from gitlab
	if c.Project, err = (Relation{}).FetchRalation(*serviceName); err != nil {
		return
	}
	if err = writeConfigYml(c); err != nil {
		return
	}
	if err = writeNgnix(c.Project, c.Port.EventBroker); err != nil {
		return
	}
	return
}

const (
	ngnixServer = `server {
		listen       80;
		server_name  test.local.com;
		location / {
			root   /usr/share/nginx/html;
			index  index.html index.htm;
		}
		`
	ngnixLocation = `location /$serverName/ {
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_set_header Connection keep-alive;
		proxy_pass       http://$containerName:$port/;
	}
	`
)

func getNginxLocation(serverName, port string) (location string) {
	location = strings.Replace(ngnixLocation, "$serverName", serverName, -1)
	location = strings.Replace(location, "$containerName", Compose{}.getContainerName(serverName), -1)
	location = strings.Replace(location, "$port", port, -1)
	return
}

func getContainerPort(port string) (containerPort string) {
	containerPort = port[strings.LastIndex(port, ":")+1:]
	return
}

// setNgnix set nginx default.conf
func writeNgnix(p *ProjectDto, eventBrokerPort string) (err error) {
	var location string
	location += getNginxLocation(p.ServiceName, getContainerPort(p.Ports[0]))

	for _, sp := range p.SubProjects {
		location += getNginxLocation(sp.ServiceName, getContainerPort(sp.Ports[0]))
	}

	if shouldStartEventBroker(p) {
		location += getNginxLocation(EventBroker_Name, eventBrokerPort)
	}
	if err = os.MkdirAll(TEMP_FILE+"/nginx", os.ModePerm); err != nil {
		return
	}
	return writeFile(TEMP_FILE+"/nginx/default.conf", ngnixServer+location+"\n}")
}

func testProjectDependency(serviceName string) (projectDto *ProjectDto, err error) {

	vip := viper.New()
	vip.AddConfigPath(".")
	vip.SetConfigName("relation")

	if err = vip.ReadInConfig(); err != nil {
		err = fmt.Errorf("Fatal error config file: %s \n", err)
		return
	}
	projectDto = &ProjectDto{}
	if err = vip.Unmarshal(projectDto); err != nil {
		err = fmt.Errorf("Fatal error config file: %s \n", err)
		return
	}
	return

}

func loadEnv(c *ConfigDto, scope string, ip, appEnv,
	mysqlPort, redisPort, mongoPort, sqlServerPort, kafkaPort, kafkaSecondPort, eventBrokerPort, nginxPort, zookeeperPort *string) (err error) {

	if appEnv != nil && len(*appEnv) != 0 {
		app_env = *appEnv
	} else {
		app_env = "qa"
	}

	c.Ip, err = getIp(ip)
	if err != nil {
		return
	}
	c.Scope = scope
	if c.Project == nil {
		c.Project = &ProjectDto{}
	}
	if mysqlPort == nil || len(*mysqlPort) == 0 {
		c.Port.Mysql = inPort.Mysql
	} else {
		c.Port.Mysql = *mysqlPort
	}
	if redisPort == nil || len(*redisPort) == 0 {
		c.Port.Redis = inPort.Redis
	} else {
		c.Port.Mysql = *mysqlPort
	}
	if mongoPort == nil || len(*mongoPort) == 0 {
		c.Port.Mongo = inPort.Mongo
	} else {
		c.Port.Mongo = *mongoPort
	}
	if sqlServerPort == nil || len(*sqlServerPort) == 0 {
		c.Port.SqlServer = inPort.SqlServer
	} else {
		c.Port.SqlServer = *sqlServerPort
	}
	if kafkaPort == nil || len(*kafkaPort) == 0 {
		c.Port.Kafka = inPort.Kafka
	} else {
		c.Port.Kafka = *kafkaPort
	}
	if kafkaSecondPort == nil || len(*kafkaSecondPort) == 0 {
		c.Port.KafkaSecond = inPort.KafkaSecond
	} else {
		c.Port.KafkaSecond = *kafkaSecondPort
	}
	if zookeeperPort == nil || len(*zookeeperPort) == 0 {
		c.Port.Zookeeper = inPort.Zookeeper
	} else {
		c.Port.Zookeeper = *zookeeperPort
	}

	if eventBrokerPort == nil || len(*eventBrokerPort) == 0 {
		c.Port.EventBroker = inPort.EventBroker
	} else {
		c.Port.EventBroker = *eventBrokerPort
	}
	// nginx default outPort:3001
	if nginxPort == nil || len(*nginxPort) == 0 {
		c.Port.Nginx = "3001"
	} else {
		c.Port.Nginx = *nginxPort
	}

	return
}

func writeConfigYml(c *ConfigDto) (err error) {
	vip := viper.New()
	vip.SetConfigName(YMLNAMECONFIG)
	vip.AddConfigPath(TEMP_FILE)
	vip.Set("scope", c.Scope)
	vip.Set("port", c.Port)
	vip.Set("project", c.Project)
	err = writeConfig(TEMP_FILE+"/"+YMLNAMECONFIG+".yml", vip)
	if err != nil {
		err = fmt.Errorf("write to config.yml error:%v", err)
		return
	}
	return
}

func getScope(updated *string) (updatedStr string, err error) {
	if updated == nil || len(*updated) == 0 {
		updatedStr = LOCAL.String()
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
	return
}

func writeConfig(path string, viper *viper.Viper) (err error) {
	if err = os.MkdirAll(TEMP_FILE, os.ModePerm); err != nil {
		return
	}
	if err = createIfNot(path); err != nil {
		return
	}
	if err = viper.WriteConfig(); err != nil {
		return
	}
	return
}

func scopeSettings(scope string) (isLocalConfig bool) {
	if _, err := os.Stat(TEMP_FILE + "/" + YMLNAMECONFIG + ".yml"); err != nil {
		isLocalConfig = false
	} else {
		if scope == LOCAL.String() {
			isLocalConfig = true
		}
	}
	updatedConfig = scope
	return
}
