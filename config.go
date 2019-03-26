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
	if c.Project, err = testProjectDependency(c.Project.GitShortPath); err != nil {
		return
	}
	if err = loadProjectEnv(c.Project); err != nil {
		return
	}
	setConfigEnv(c)

	if err = writeConfigYml(c); err != nil {
		return
	}
	return
}

func testProjectDependency(gitShortPath string) (projectDto *ProjectDto, err error) {
	// for _, projectDto := range c.Project.SubProjects {
	// 	c.Project.SubNames = append(c.Project.SubNames, projectDto.Name)
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
			err = fmt.Errorf("parse project.yml error,project:%v,err:%v", subProject.Name, err.Error())
			return
		}
		setPort(projectDto.SubProjects[i])
	}
	return
}

func setPort(projectDto *ProjectDto) {
	ports, err := freeport.GetFreePorts(len(projectDto.Ports))
	if err != nil {
		err = fmt.Errorf("get free port error,project:%v,err:%v", projectDto.Name, err.Error())
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
