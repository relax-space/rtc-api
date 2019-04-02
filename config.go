package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/phayes/freeport"
	"github.com/spf13/viper"
)

func LoadEnv() (c *ConfigDto, err error) {
	serviceName := flag.String("serviceName", os.Getenv("serviceName"), "serviceName")
	updated := flag.String("updated", os.Getenv("updated"), "updated")
	mysqlPort := flag.String("mysqlPort", os.Getenv("mysqlPort"), "mysqlPort")
	redisPort := flag.String("redisPort", os.Getenv("redisPort"), "redisPort")
	mongoPort := flag.String("mongoPort", os.Getenv("mongoPort"), "mongoPort")
	sqlServerPort := flag.String("sqlServerPort", os.Getenv("sqlServerPort"), "sqlServerPort")
	kafkaPort := flag.String("kafkaPort", os.Getenv("kafkaPort"), "kafkaPort")

	flag.Parse()

	updatedStr, err := getScope(updated)
	if err != nil {
		err = fmt.Errorf("read env error:%v", err)
		return
	}

	c = &ConfigDto{}
	if err = loadEnv(c, updatedStr, serviceName,
		mysqlPort, redisPort, mongoPort, sqlServerPort, kafkaPort); err != nil {
		return
	}
	isLocalConfig := shouldLocalConfig(updatedStr)
	if isLocalConfig {
		if err = Read("", c); err != nil {
			err = fmt.Errorf("read config error:%v", err)
			return
		}
		return
	}

	//1.load base info from gitlab
	if c.Project, err = testProjectDependency(c.Project.ServiceName); err != nil {
		return
	}
	if err = loadProjectEnv(c.Project); err != nil {
		return
	}
	if err = writeConfigYml(c); err != nil {
		return
	}
	if err = writeNgnix(c.Project); err != nil {
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
		proxy_pass       http://test-$serverName:$port/;
	}
	`
)

func getNginxLocation(serverName, port string) (location string) {
	location = strings.Replace(ngnixLocation, "$serverName", serverName, -1)
	location = strings.Replace(location, "$port", port, -1)
	return
}

func getContainerPort(port string) (containerPort string) {
	containerPort = port[strings.LastIndex(port, ":")+1:]
	return
}

// setNgnix set nginx default.conf
func writeNgnix(p *ProjectDto) (err error) {
	var location string
	location += getNginxLocation(p.ServiceName, getContainerPort(p.Ports[0]))

	for _, sp := range p.SubProjects {
		location += getNginxLocation(sp.ServiceName, getContainerPort(sp.Ports[0]))
	}

	return writeFile("default.conf", ngnixServer+location+"\n}")
}

func testProjectDependency(serviceName string) (projectDto *ProjectDto, err error) {
	// for _, projectDto := range c.Project.SubProjects {
	// 	c.Project.SubNames = append(c.Project.SubNames, projectDto.ServiceName)
	// }

	// lastIndex := strings.LastIndex(gitShortPath, "/")
	// pName := gitShortPath[lastIndex:]

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

func loadEnv(c *ConfigDto, scope string,
	serviceName, mysqlPort, redisPort, mongoPort, sqlServerPort, kafkaPort *string) (err error) {
	if serviceName == nil || len(*serviceName) == 0 {
		err = fmt.Errorf("read env error:%v", "serviceName is required.")
		return
	}
	c.Scope = scope
	if c.Project == nil {
		c.Project = &ProjectDto{}
	}
	c.Project.ServiceName = *serviceName
	if mysqlPort == nil || len(*mysqlPort) == 0 {
		c.Port.Mysql = "3306"
	}
	if redisPort == nil || len(*redisPort) == 0 {
		c.Port.Redis = "6379"
	}
	if mongoPort == nil || len(*mongoPort) == 0 {
		c.Port.Mongo = "27017"
	}
	if sqlServerPort == nil || len(*sqlServerPort) == 0 {
		c.Port.SqlServer = "1433"
	}
	if kafkaPort == nil || len(*kafkaPort) == 0 {
		c.Port.Kafka = "9092"
	}
	return
}

func writeConfigYml(c *ConfigDto) (err error) {
	vip := viper.New()
	vip.SetConfigName(YmlNameConfig)
	vip.AddConfigPath(".")
	vip.Set("scope", c.Scope)
	vip.Set("port", c.Port)
	vip.Set("project", c.Project)
	err = writeConfig(YmlNameConfig+".yml", vip)
	if err != nil {
		err = fmt.Errorf("write to config.yml error:%v", err)
		return
	}
	return
}

func getScope(updated *string) (updatedStr string, err error) {
	if updated == nil || len(*updated) == 0 {
		updatedStr = NONE.String()
		return
	}
	for _, s := range NONE.List() {
		if strings.ToLower(*updated) == s {
			updatedStr = s
			break
		}
	}
	if len(updatedStr) == 0 {
		err = fmt.Errorf("Parameters(%v) are not supported, only support all, sql, app", *updated)
		return
	}
	return
}

func fetchsqlTofile(c *ConfigDto) (err error) {
	folder := ""
	if c.Project.IsComplex {
		folder = "/" + c.Project.ServiceName
	}
	urlString := fmt.Sprintf("%v/test_info%v/table.sql", c.Project.GitRaw, folder)
	if err = fetchTofile(urlString, c.Project.ServiceName+".sql", PrivateToken); err != nil {
		err = fmt.Errorf("read table.sql error:%v", err)
		return
	}
	for _, projectDto := range c.Project.SubProjects {
		urlString := fmt.Sprintf("%v/test_info%v/table.sql", projectDto.GitRaw, folder)
		if err = fetchTofile(urlString, projectDto.ServiceName+".sql", PrivateToken); err != nil {
			err = fmt.Errorf("read %v.sql error:%v", projectDto.ServiceName, err)
			return
		}
	}
	return
}

func writeConfig(path string, viper *viper.Viper) (err error) {
	if err = createIfNot(path); err != nil {
		return
	}
	if err = viper.WriteConfig(); err != nil {
		return
	}
	return
}

func loadProjectEnv(projectDto *ProjectDto) (err error) {

	projectName := projectDto.ServiceName
	projectDto.GitRaw = fmt.Sprintf("%v/%v/raw/qa", PreGitHttpUrl, projectDto.GitShortPath)
	urlString := projectDto.GitRaw + "/test_info/project.yml"
	b, err := fetchFromgitlab(urlString, PrivateToken)
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectName, err.Error())
		return
	}
	setPort(projectDto)

	for i, subProject := range projectDto.SubProjects {
		projectDto.SubProjects[i].GitRaw = fmt.Sprintf("%v/%v/raw/qa", PreGitHttpUrl, subProject.GitShortPath)
		urlString := subProject.GitRaw + "/test_info/project.yml"
		b, err = fetchFromgitlab(urlString, PrivateToken)
		if err = yaml.Unmarshal(b, projectDto.SubProjects[i]); err != nil {
			err = fmt.Errorf("parse project.yml error,project:%v,err:%v", subProject.ServiceName, err.Error())
			return
		}
		setPort(projectDto.SubProjects[i])
	}
	return
}

func setPort(projectDto *ProjectDto) {
	ports, err := freeport.GetFreePorts(len(projectDto.Ports))
	if err != nil {
		err = fmt.Errorf("get free port error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	for i, _ := range projectDto.Ports {
		projectDto.Ports[i] = fmt.Sprintf("%v:%v", ports[i], projectDto.Ports[i])
	}
}

func shouldLocalConfig(scope string) (isLocalConfig bool) {
	if _, err := os.Stat(YmlNameConfig + ".yml"); err != nil {
		isLocalConfig = false
	} else {
		if scope == NONE.String() {
			isLocalConfig = true
		}
	}
	return
}

func shouldUpdateData(scope string) bool {

	return scope == ALL.String() || scope == DATA.String()
}
func shouldUpdateCompose(scope string) bool {
	if _, err := os.Stat(YmlNameDockerCompose + ".yml"); err != nil {
		return true
	}
	return scope != NONE.String()
}
func shouldUpdateApp(scope string) bool {
	return scope == ALL.String() || scope == DATA.String()
}

func shouldRestartData(scope string) bool {
	return scope == ALL.String() || scope == DATA.String()
}

func shouldRestartApp(scope string) bool {
	return scope == ALL.String() || scope == DATA.String()
}

func shouldStartKakfa(project *ProjectDto) (isKafka bool) {
	if project.IsProjectKafka {
		isKafka = true
	} else {
		for _, subProject := range project.SubProjects {
			if subProject.IsProjectKafka {
				isKafka = true
				break
			}
		}
	}
	return
}

func shouldStartMysql(project *ProjectDto) bool {
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == MYSQL.String() {
			return true
		}
	}
	return false
}

func shouldStartRedis(project *ProjectDto) bool {
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == REDIS.String() {
			return true
		}
	}
	return false
}

func shouldStartMongo(project *ProjectDto) bool {
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == MONGO.String() {
			return true
		}
	}
	return false
}

func shouldStartSqlServer(project *ProjectDto) bool {
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == SQLSERVER.String() {
			return true
		}
	}
	return false
}

func databaseList(project *ProjectDto) (list map[string]string) {
	list = make(map[string]string, 0)
	for _, d := range project.Databases {
		list[d] = d
	}
	for _, subProject := range project.SubProjects {
		for _, d := range subProject.Databases {
			list[d] = d
		}
	}
	return
}
