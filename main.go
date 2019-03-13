package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/ghodss/yaml"

	"github.com/pangpanglabs/goutils/httpreq"

	"github.com/matishsiao/goInfo"

	"github.com/spf13/viper"
)

const (
	Windows      = "windows"
	Linux        = "linux"
	PrivateToken = "Su5_HzvQxtyANyDtzx_P"
)

type ConfigDto struct {
	IsKafka bool `json:"isKafka"`
	Mysql   struct {
		Envs []string `json:"envs"`
	} `json:"mysql"`
	Project      *ProjectDto `json:"project"`
	Gopath       string      `json:"gopath"`
	GopathDocker string      `json:"gopathDocker"`
}
type ProjectDto struct {
	Name        string        `json:"name"`
	Envs        []string      `json:"envs"`
	Ports       []string      `json:"ports"`
	Databases   []string      `json:"database"`
	SubNames    []string      `json:"subNames"`
	SubProjects []*ProjectDto `json:"subProjects"`
}

func main() {
	var c ConfigDto
	err := Read("", &c)
	if err != nil {
		fmt.Printf("read config error:%v", err)
		return
	}
	//1.load base info from gitlab
	loadProjectEnv(c.Project)
	setMysqlEnv(&c)

	if err := fetchsqlTofile(c); err != nil {
		fmt.Println(err)
		return
	}

	//2. generate docker-compose
	viper := viper.New()
	setComposeMysql(viper, c.Mysql.Envs)
	setComposeKafka(viper)

	for _, projectDto := range c.Project.SubProjects {
		c.Project.SubNames = append(c.Project.SubNames, projectDto.Name)
	}

	if err = AppCompose(viper, c.Gopath, c.Project, c.Project.SubProjects); err != nil {
		fmt.Printf("write to app error:%v", err)
		return
	}
	//3. run docker-compose
	if _, err = Cmd("docker-compose down"); err != nil {
		fmt.Printf("err:%v", err)
	}
	fmt.Println("==> compose downed!")
	if _, err = Cmd("docker-compose build"); err != nil {
		fmt.Printf("err:%v", err)
	}
	fmt.Println("==> compose builded!")

	if _, err = Cmd("docker-compose up"); err != nil {
		fmt.Printf("err:%v", err)
	}
	fmt.Println("==> compose up!")

}

func setMysqlEnv(c *ConfigDto) {
	dbNames := make([]string, 0)
	for _, db := range c.Project.Databases {
		dbNames = append(dbNames, db)
	}

	for _, subProject := range c.Project.SubProjects {
		for _, db := range subProject.Databases {
			dbNames = append(dbNames, db)
		}
	}
	c.Mysql.Envs = append(c.Mysql.Envs, "MYSQL_ROOT_PASSWORD=1234")
	for i, name := range dbNames {
		c.Mysql.Envs = append(c.Mysql.Envs, fmt.Sprintf("MYSQL_DATABASE_%v=%v", i+1, name))
	}
}

func loadProjectEnv(projectDto *ProjectDto) {
	projectName := projectDto.Name
	urlString := fmt.Sprintf("https://gitlab.p2shop.cn:8443/data/test-db/raw/master/%v/project.yml", projectName)
	b, err := fetchFromgitlab(urlString, PrivateToken)
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		fmt.Printf("parse yaml error,project:%v,err:%v", projectName, err.Error())
		return
	}
	for i, subProject := range projectDto.SubProjects {
		urlString := fmt.Sprintf("https://gitlab.p2shop.cn:8443/data/test-db/raw/master/%v/project.yml", subProject.Name)
		b, err := fetchFromgitlab(urlString, PrivateToken)

		if err = yaml.Unmarshal(b, projectDto.SubProjects[i]); err != nil {
			fmt.Printf("parse yaml error,project:%v,err:%v", subProject.Name, err.Error())
			return
		}
	}
}

func fetchsqlTofile(c ConfigDto) (err error) {
	projectName := c.Project.Name
	urlString := fmt.Sprintf("https://gitlab.p2shop.cn:8443/data/test-db/raw/master/%v/table.sql", projectName)
	if err = fetchTofile(urlString, projectName+".sql", PrivateToken); err != nil {
		err = fmt.Errorf("read table.sql error:%v", err)
		return
	}

	for _, projectDto := range c.Project.SubProjects {
		urlString := fmt.Sprintf("https://gitlab.p2shop.cn:8443/data/test-db/raw/master/%v/table.sql", projectDto.Name)
		if err = fetchTofile(urlString, projectDto.Name+".sql", PrivateToken); err != nil {
			err = fmt.Errorf("read %v.sql error:%v", projectDto.Name, err)
			return
		}
	}
	return
}

func fetchFromgitlab(url, privateToken string) (b []byte, err error) {
	req := httpreq.New(http.MethodGet, url, nil, func(httpReq *httpreq.HttpReq) error {
		httpReq.RespDataType = httpreq.ByteArrayType
		return nil
	})
	req.Req.Header.Set("PRIVATE-TOKEN", privateToken)
	fmt.Println(req.Req.Header)
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
	fmt.Println(req.Req.Header)
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

func AppCompose(viper *viper.Viper, gopath string, project *ProjectDto, subProjects []*ProjectDto) error {
	appComposeMain(viper, gopath, project)
	for _, project := range subProjects {
		appCompose(viper, gopath, project)
	}
	//viper.Set("networks.sharenet.driver", "bridge")
	viper.Set("version", "3")
	return viper.WriteConfig()
}

func appComposeMain(viper *viper.Viper, gopath string, project *ProjectDto) {
	servicePre := "services." + project.Name

	viper.SetConfigName("docker-compose")
	viper.AddConfigPath(".")

	project.SubNames = append(project.SubNames, "kafkaserver")
	project.SubNames = append(project.SubNames, "mysqlserver")
	viper.Set(servicePre+".build.context", gopath+"/src/"+project.Name)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", "test-"+project.Name)
	viper.Set(servicePre+".restart", "on-failure:5")

	viper.Set(servicePre+".container_name", "test-"+project.Name)
	viper.Set(servicePre+".depends_on", project.SubNames)
	//	viper.Set(servicePre+".networks", []string{"sharenet"})
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

//env format []string{"MYSQL_ROOT_PASSWORD=1234"}
func appCompose(viper *viper.Viper, gopath string, project *ProjectDto) {
	servicePre := "services." + project.Name
	viper.Set(servicePre+".build.context", gopath+"/src/"+project.Name)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", "test-"+project.Name)
	viper.Set(servicePre+".restart", "on-failure:5")

	viper.Set(servicePre+".depends_on", []string{"mysqlserver"})
	viper.Set(servicePre+".container_name", "test-"+project.Name)
	//	viper.Set(servicePre+".networks", []string{"sharenet"})
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

func setComposeMysql(viper *viper.Viper, envs []string) {
	viper.Set("version", "3")
	viper.Set("services.mysqlserver.image", "gruppometasrl/mysql57")
	viper.Set("services.mysqlserver.container_name", "test-mysql")
	viper.Set("services.mysqlserver.volumes", []string{".:/docker-entrypoint-initdb.d"})
	viper.Set("services.mysqlserver.ports", []string{"3306:3306"})
	viper.Set("services.mysqlserver.restart", "always")
	//viper.Set("services.mysqlserver.networks", []string{"sharenet"})
	viper.Set("services.mysqlserver.environment", envs)
}

func setComposeKafka(viper *viper.Viper) {
	viper.Set("services.kafkaserver.image", "spotify/kafka:latest")
	viper.Set("services.kafkaserver.container_name", "test-kafka")
	viper.Set("services.kafkaserver.hostname", "kafkaserver")
	viper.Set("services.kafkaserver.restart", "always")
	//viper.Set("services.kafkaserver.networks", []string{"sharenet"})
	viper.Set("services.kafkaserver.ports", []string{"2181:2181", "9092:9092"})
	viper.Set("services.kafkaserver.environment", []string{"ADVERTISED_HOST=kafkaserver",
		"ADVERTISED_PORT=9092"})
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
