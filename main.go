package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ghodss/yaml"

	"github.com/pangpanglabs/goutils/httpreq"

	"github.com/matishsiao/goInfo"

	"github.com/spf13/viper"
)

const (
	Windows              = "windows"
	Linux                = "linux"
	PrivateToken         = "Su5_HzvQxtyANyDtzx_P"
	PreGitSshUrl         = "ssh://git@gitlab.p2shop.cn:822"
	PreGitHttpUrl        = "https://gitlab.p2shop.cn:8443"
	YmlNameConfig        = "config"
	YmlNameDockerCompose = "docker-compose"
)
const (
	ScopeALL  = "ALL"
	ScopeData = "DATA"
	ScopeAPP  = "APP"
	ScopeNONE = "NONE"
)

var (
	scopes = []string{ScopeALL, ScopeData, ScopeAPP, ScopeNONE}
)

type ConfigDto struct {
	Scope   string
	NoCache bool
	IsKafka bool
	Mysql   struct {
		Databases []string
		Ports     []string
	}
	Project *ProjectDto
}
type ProjectDto struct {
	Name           string //eg. ipay-api
	GitShortPath   string //eg. ipay/ipay-api
	GitRaw         string
	Envs           []string // from jenkins
	IsProjectKafka bool
	Ports          []string
	Databases      []string
	SubNames       []string
	SubProjects    []*ProjectDto
	//Dependencies []string//delete
}

func main() {

	c, err := LoadEnv()
	if err != nil {
		fmt.Println(err)
		return
	}

	//1.download sql data
	if shouldUpdateData(c.Scope) {
		if err := fetchsqlTofile(c); err != nil {
			fmt.Println(err)
			return
		}
	}

	//2. generate docker-compose
	if shouldUpdateCompose(c.Scope) {
		viper := viper.New()
		if c.IsKafka {
			setComposeKafka(viper)
		}
		if shouldStartMysql(c.Mysql.Databases) {
			setComposeMysql(viper, c.Mysql.Ports, c.Mysql.Databases)
		}
		setComposeApp(viper, c.Project)

		if err = writeConfig(YmlNameDockerCompose+".yml", viper); err != nil {
			fmt.Printf("write to config.yml error:%v", err)
			return
		}
	}

	//3. run docker-compose
	if shouldRestartData(c.Scope, c.NoCache) {
		if _, err = Cmd("docker-compose -f docker-compose.yml down"); err != nil {
			fmt.Printf("err:%v", err)
		}
		fmt.Println("==> compose downed!")
	}

	if shouldRestartApp(c.Scope, c.NoCache) {
		if _, err = Cmd("docker-compose -f docker-compose.yml build"); err != nil {
			fmt.Printf("err:%v", err)
		}
		fmt.Println("==> compose builded!")
	}

	go func() {
		if _, err = Cmd("docker-compose -f docker-compose.yml up"); err != nil {
			fmt.Printf("err:%v", err)
		}
	}()
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
	time.Sleep(10 * time.Second)
	fmt.Println("==> compose may have started !")
	time.Sleep(10 * time.Minute)
}

func shouldLocalConfig(scope string) (isLocalConfig bool) {
	if _, err := os.Stat(YmlNameConfig + ".yml"); err != nil {
		isLocalConfig = false
	} else {
		if scope == ScopeNONE {
			isLocalConfig = true
		}
	}
	return
}

func shouldStartMysql(databases []string) (isStart bool) {
	if len(databases) != 0 {
		isStart = true
		return
	}
	return
}

func shouldUpdateData(scope string) bool {

	return scope == ScopeALL || scope == ScopeData
}
func shouldUpdateCompose(scope string) bool {
	if _, err := os.Stat(YmlNameDockerCompose + ".yml"); err != nil {
		return true
	}
	return scope != ScopeNONE
}
func shouldUpdateApp(scope string) bool {
	return scope == ScopeALL || scope == ScopeData
}

func shouldRestartData(scope string, noCache bool) bool {
	if noCache {
		return true
	}
	return scope == ScopeALL || scope == ScopeData
}

func shouldRestartApp(scope string, noCache bool) bool {
	if noCache {
		return true
	}
	return scope == ScopeALL || scope == ScopeData
}

// load base info

func testProjectDependency(gitShortPath string) *ProjectDto {
	// for _, projectDto := range c.Project.SubProjects {
	// 	c.Project.SubNames = append(c.Project.SubNames, projectDto.Name)
	// }

	// lastIndex := strings.LastIndex(gitShortPath, "/")
	// pName := gitShortPath[lastIndex:]
	return &ProjectDto{
		Name:         "ibill-api",
		GitShortPath: "ipay/ibill-api",
		SubNames:     []string{"ipay-api"},
		SubProjects: []*ProjectDto{
			// &ProjectDto{
			// 	Name:         "pay-api",
			// 	GitShortPath: "omni/pay-api",
			// },
			&ProjectDto{
				Name:         "ipay-api",
				GitShortPath: "ipay/ipay-api",
			},
		},
	}
}

func loadEnv(c *ConfigDto, scope, gitShortPath string, mysqlPorts []string, noCache bool) {
	c.Scope = scope
	c.NoCache = noCache
	c.Mysql.Ports = mysqlPorts
	if c.Project == nil {
		c.Project = &ProjectDto{}
	}
	c.Project.GitShortPath = gitShortPath

}

func writeConfigYml(c *ConfigDto) (err error) {
	vip := viper.New()
	vip.SetConfigName(YmlNameConfig)
	vip.AddConfigPath(".")
	vip.Set("scope", c.Scope)
	vip.Set("isKafka", c.IsKafka)
	vip.Set("mysql", c.Mysql)
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
		updatedStr = ScopeNONE
		return
	}
	for _, s := range scopes {
		if strings.ToUpper(*updated) == s {
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
func LoadEnv() (c *ConfigDto, err error) {
	gitShortPath := flag.String("gitShortPath", os.Getenv("gitShortPath"), "gitShortPath")
	updated := flag.String("updated", os.Getenv("updated"), "updated")
	mysqlPort := flag.String("mysqlPort", os.Getenv("mysqlport"), "mysqlPort")
	noCache := flag.String("no-cache", os.Getenv("no-cache"), "no-cache")

	flag.Parse()

	if gitShortPath == nil || len(*gitShortPath) == 0 {
		err = fmt.Errorf("read env error:%v", "gitShortPath is required.")
		return
	}

	noCahceBool := true
	if noCache == nil || len(*noCache) == 0 {
		noCahceBool = false
	}

	updatedStr, err := getScope(updated)
	if err != nil {
		err = fmt.Errorf("read env error:%v", err)
		return
	}

	var mysqlPorts []string
	if mysqlPort == nil || len(*mysqlPort) == 0 {
		mysqlPorts = append(mysqlPorts, "3306:3306")
	}
	shortPath := *gitShortPath
	c = &ConfigDto{}
	isLocalConfig := shouldLocalConfig(updatedStr)
	if isLocalConfig {
		if err = Read("", c); err != nil {
			err = fmt.Errorf("read config error:%v", err)
			return
		}
		loadEnv(c, updatedStr, shortPath, mysqlPorts, noCahceBool)
		return
	}
	loadEnv(c, updatedStr, shortPath, mysqlPorts, noCahceBool)

	//1.load base info from gitlab
	c.Project = testProjectDependency(c.Project.GitShortPath)
	if err = loadProjectEnv(c.Project); err != nil {
		return
	}
	setConfigEnv(c)

	if err = writeConfigYml(c); err != nil {
		return
	}
	return
}

func setConfigEnv(c *ConfigDto) {
	dbNames := make(map[string]string, 0)

	var isKafka bool
	if c.Project.IsProjectKafka {
		isKafka = true
	} else {
		for _, subProject := range c.Project.SubProjects {
			if subProject.IsProjectKafka {
				isKafka = true
				break
			}
		}
	}
	c.IsKafka = isKafka

	for _, db := range c.Project.Databases {
		dbNames[db] = db
	}

	for _, subProject := range c.Project.SubProjects {
		for _, db := range subProject.Databases {
			dbNames[db] = db
		}
	}
	var index int
	for _, name := range dbNames {
		index++
		c.Mysql.Databases = append(c.Mysql.Databases, fmt.Sprintf("MYSQL_DATABASE_%v=%v", index, name))
	}
}

func loadProjectEnv(projectDto *ProjectDto) (err error) {

	projectName := projectDto.Name
	projectDto.GitRaw = fmt.Sprintf("%v/%v/raw/qa", PreGitHttpUrl, projectDto.GitShortPath)
	urlString := projectDto.GitRaw + "/test_info/project.yml"
	fmt.Println(urlString)
	b, err := fetchFromgitlab(urlString, PrivateToken)
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectName, err.Error())
		return
	}

	for i, subProject := range projectDto.SubProjects {
		projectDto.SubProjects[i].GitRaw = fmt.Sprintf("%v/%v/raw/qa", PreGitHttpUrl, subProject.GitShortPath)
		urlString := subProject.GitRaw + "/test_info/project.yml"
		b, err = fetchFromgitlab(urlString, PrivateToken)
		if err = yaml.Unmarshal(b, projectDto.SubProjects[i]); err != nil {
			err = fmt.Errorf("parse project.yml error,project:%v,err:%v", subProject.Name, err.Error())
			return
		}
	}
	return
}

func fetchsqlTofile(c *ConfigDto) (err error) {
	urlString := c.Project.GitRaw + "/test_info/table.sql"
	if err = fetchTofile(urlString, c.Project.Name+".sql", PrivateToken); err != nil {
		err = fmt.Errorf("read table.sql error:%v", err)
		return
	}
	for _, projectDto := range c.Project.SubProjects {
		urlString := projectDto.GitRaw + "/test_info/table.sql"
		if err = fetchTofile(urlString, projectDto.Name+".sql", PrivateToken); err != nil {
			err = fmt.Errorf("read %v.sql error:%v", projectDto.Name, err)
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

// generate docker-compose
func setComposeApp(viper *viper.Viper, project *ProjectDto) {
	appComposeMain(viper, project)
	for _, project := range project.SubProjects {
		appCompose(viper, project)
	}
	viper.Set("version", "3")
}

func appComposeMain(viper *viper.Viper, project *ProjectDto) {
	servicePre := "services." + project.Name

	viper.SetConfigName(YmlNameDockerCompose)
	viper.AddConfigPath(".")

	project.SubNames = append(project.SubNames, "kafkaserver")
	project.SubNames = append(project.SubNames, "mysqlserver")
	viper.Set(servicePre+".build.context", os.Getenv("GOPATH")+"/src/"+project.Name)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", "test-"+project.Name)
	viper.Set(servicePre+".restart", "on-failure:5")

	viper.Set(servicePre+".container_name", "test-"+project.Name)
	viper.Set(servicePre+".depends_on", project.SubNames)
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

//env format []string{"MYSQL_ROOT_PASSWORD=1234"}
func appCompose(viper *viper.Viper, project *ProjectDto) {
	servicePre := "services." + project.Name
	viper.Set(servicePre+".build.context", os.Getenv("GOPATH")+"/src/"+project.Name)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", "test-"+project.Name)
	viper.Set(servicePre+".restart", "on-failure:5")

	viper.Set(servicePre+".depends_on", []string{"mysqlserver"})
	viper.Set(servicePre+".container_name", "test-"+project.Name)
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

func setComposeMysql(viper *viper.Viper, ports, envs []string) {
	envs = append(envs, "MYSQL_ROOT_PASSWORD=1234")
	viper.Set("services.mysqlserver.image", "gruppometasrl/mysql57")
	viper.Set("services.mysqlserver.container_name", "test-mysql")
	viper.Set("services.mysqlserver.volumes", []string{".:/docker-entrypoint-initdb.d"})
	viper.Set("services.mysqlserver.ports", ports)
	viper.Set("services.mysqlserver.restart", "always")
	viper.Set("services.mysqlserver.environment", envs)
}

func setComposeKafka(viper *viper.Viper) {
	viper.Set("services.kafkaserver.image", "spotify/kafka:latest")
	viper.Set("services.kafkaserver.container_name", "test-kafka")
	viper.Set("services.kafkaserver.hostname", "kafkaserver")
	viper.Set("services.kafkaserver.restart", "always")
	viper.Set("services.kafkaserver.ports", []string{"2181:2181", "9092:9092"})
	viper.Set("services.kafkaserver.environment", []string{"ADVERTISED_HOST=kafkaserver",
		"ADVERTISED_PORT=9092"})
}

// util

func Cmd(cmd string) (out []byte, err error) {
	if goInfo.GetInfo().GoOS == Windows {
		out, err = exec.Command("cmd", "/c", cmd).Output()
	} else {
		out, err = exec.Command("bash", "-c", cmd).Output()
	}
	if err != nil {
		panic(fmt.Sprintf("some error found:%v,detail:%v", err.Error(), string(out)))
	}
	return
}

func createIfNot(path string) error {
	if _, err := os.Stat(path); err != nil {
		if _, err = os.Create(path); err != nil {
			return err
		}
	}
	return nil
}

func Read(env string, config interface{}) error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}

	if env != "" {
		f, err := os.Open("config." + env + ".yml")
		if err != nil {
			return fmt.Errorf("Fatal error config file: %s \n", err)
		}
		defer f.Close()
		viper.MergeConfig(f)
	}

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}
	return nil
}

func fetchFromgitlab(url, privateToken string) (b []byte, err error) {
	req := httpreq.New(http.MethodGet, url, nil, func(httpReq *httpreq.HttpReq) error {
		httpReq.RespDataType = httpreq.ByteArrayType
		return nil
	})
	req.Req.Header.Set("PRIVATE-TOKEN", privateToken)
	resp, err := req.RawCall()
	defer resp.Body.Close()
	if err != nil {
		return
	}
	b, err = ioutil.ReadAll(resp.Body)
	return
}

func fetchTofile(url, fileName, privateToken string) (err error) {
	req := httpreq.New(http.MethodGet, url, nil, func(httpReq *httpreq.HttpReq) error {
		httpReq.RespDataType = httpreq.ByteArrayType
		return nil
	})
	req.Req.Header.Set("PRIVATE-TOKEN", privateToken)
	resp, err := req.RawCall()
	defer resp.Body.Close()
	if err != nil {
		return
	}

	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return
}

// func getFreePort() (port int, err error) {
// 	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
// 	if err != nil {
// 		return
// 	}
// 	port = addr.Port
// 	return
// }

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
