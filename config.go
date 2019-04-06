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
	eventBrokerPort := flag.String("eventBrokerPort", os.Getenv("eventBrokerPort"), "eventBrokerPort")
	nginxPort := flag.String("nginxPort", os.Getenv("nginxPort"), "nginxPort")

	//EventBroker
	flag.Parse()

	updatedStr, err := getScope(updated)
	if err != nil {
		err = fmt.Errorf("read env error:%v", err)
		return
	}

	c = &ConfigDto{}
	if err = loadEnv(c, updatedStr, serviceName,
		mysqlPort, redisPort, mongoPort, sqlServerPort, kafkaPort, eventBrokerPort, nginxPort); err != nil {
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
	serviceName, mysqlPort, redisPort, mongoPort, sqlServerPort, kafkaPort, eventBrokerPort, nginxPort *string) (err error) {
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
	if eventBrokerPort == nil || len(*eventBrokerPort) == 0 {
		c.Port.EventBroker = "3000"
	}
	if nginxPort == nil || len(*nginxPort) == 0 {
		c.Port.Nginx = "3001"
	}
	return
}

func writeConfigYml(c *ConfigDto) (err error) {
	vip := viper.New()
	vip.SetConfigName(YmlNameConfig)
	vip.AddConfigPath(TEMP_FILE)
	vip.Set("scope", c.Scope)
	vip.Set("port", c.Port)
	vip.Set("project", c.Project)
	err = writeConfig(TEMP_FILE+"/"+YmlNameConfig+".yml", vip)
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
		err = fmt.Errorf("Parameters(%v) are not supported, only support %v", *updated, NONE.List())
		return
	}
	return
}

func fetchsqlTofile(project *ProjectDto) (err error) {

	path := getTestInfoPath(project)

	urlString := fmt.Sprintf("%v/test_info%v/table.sql", project.GitRaw, path)
	if err = fetchTofile(urlString,
		fmt.Sprintf("%v/%v.sql", TEMP_FILE, project.ServiceName),
		PrivateToken); err != nil {
		err = fmt.Errorf("read table.sql error:%v", err)
		return
	}
	for _, projectDto := range project.SubProjects {
		urlString := fmt.Sprintf("%v/test_info%v/table.sql", projectDto.GitRaw, path)
		if err = fetchTofile(urlString,
			fmt.Sprintf("%v/%v.sql", TEMP_FILE, projectDto.ServiceName),
			PrivateToken); err != nil {
			err = fmt.Errorf("read %v.sql error:%v", projectDto.ServiceName, err)
			return
		}
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

func getFirstProjectEnv(projectDto *ProjectDto) (err error) {
	urlString := fmt.Sprintf("%v/test_info/project.yml", projectDto.GitRaw)
	b, err := fetchFromgitlab(urlString, PrivateToken)
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	return
}

func getTestInfoPath(projectDto *ProjectDto) (path string) {

	// if len(projectDto.ParentFolderName) != 0 {
	// 	path = "/" + projectDto.ParentFolderName
	// }
	if projectDto.IsMulti {
		path += "/" + projectDto.ServiceName
	}
	return
}

func loadProjectEnv(projectDto *ProjectDto) (err error) {
	projectDto.GitRaw = fmt.Sprintf("%v/%v/raw/qa", PreGitHttpUrl, projectDto.GitShortPath)
	if err = getFirstProjectEnv(projectDto); err != nil {
		return
	}

	path := getTestInfoPath(projectDto)

	if projectDto.IsMulti {
		urlString := fmt.Sprintf("%v/test_info%v/project.yml", projectDto.GitRaw, path)
		b, errd := fetchFromgitlab(urlString, PrivateToken)
		if errd != nil {
			err = errd
			return
		}
		if err = yaml.Unmarshal(b, projectDto); err != nil {
			err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectDto.ServiceName, err.Error())
			return
		}
	}

	setPort(projectDto)

	for i, subProject := range projectDto.SubProjects {
		projectDto.SubProjects[i].GitRaw = fmt.Sprintf("%v/%v/raw/qa", PreGitHttpUrl, subProject.GitShortPath)
		urlString := fmt.Sprintf("%v/test_info%v/project.yml", subProject.GitRaw, path)
		b, errd := fetchFromgitlab(urlString, PrivateToken)
		if errd != nil {
			err = errd
			return
		}
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
	if _, err := os.Stat(TEMP_FILE + "/" + YmlNameConfig + ".yml"); err != nil {
		isLocalConfig = false
	} else {
		if scope == NONE.String() {
			isLocalConfig = true
		}
	}
	return
}

func shouldUpdateData(scope string) bool {

	return scope == ALL.String()
}
func shouldUpdateCompose(scope string) bool {
	if _, err := os.Stat(YmlNameDockerCompose + ".yml"); err != nil {
		return true
	}
	return scope != NONE.String()
}
func shouldUpdateApp(scope string) bool {
	return scope == ALL.String()
}

func shouldRestartData(scope string) bool {
	return scope == ALL.String()
}

func shouldRestartApp(scope string) bool {
	return scope == ALL.String()
}

func shouldStartKakfa(project *ProjectDto) (isKafka bool) {
	if shouldStartEventBroker(project) {
		return true
	}
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
	if shouldStartEventBroker(project) {
		return true
	}
	list := databaseList(project)
	for _, l := range list {
		if strings.ToLower(l) == MYSQL.String() {
			return true
		}
	}
	return false
}

func shouldStartRedis(project *ProjectDto) bool {
	if shouldStartEventBroker(project) {
		return true
	}
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

func shouldStartEventBroker(project *ProjectDto) bool {
	if list := streamList(project); len(list) != 0 {
		return true
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

func streamList(project *ProjectDto) (list map[string]string) {
	list = make(map[string]string, 0)
	for _, d := range project.StreamNames {
		list[d] = d
	}
	for _, subProject := range project.SubProjects {
		for _, d := range subProject.StreamNames {
			list[d] = d
		}
	}
	return
}
