package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/matishsiao/goInfo"

	"github.com/spf13/viper"
)

const (
	Windows = "windows"
	Linux   = "linux"
)

type ConfigDto struct {
	IsKafka bool
	Mysql   struct {
		IsUpdate bool
		Sqlpath  string
		Envs     []string
	}
	Project      *ProjectDto
	SubProjects  []*ProjectDto
	Gopath       string
	GopathDocker string
}
type ProjectDto struct {
	Name     string
	Envs     []string
	Ports    []string
	SubNames []string
}

func main() {
	var c ConfigDto
	err := Read("", &c)
	if err != nil {
		fmt.Printf("read config error:%v", err)
		return
	}
	viper := viper.New()
	setComposeMysql(viper, c.Mysql.Sqlpath, c.Mysql.Envs)
	setComposeKafka(viper)

	for _, projectDto := range c.SubProjects {
		c.Project.SubNames = append(c.Project.SubNames, projectDto.Name)
	}

	if err = AppCompose(viper, c.Gopath, c.Project, c.SubProjects); err != nil {
		fmt.Printf("write to app error:%v", err)
		return
	}
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

func setComposeMysql(viper *viper.Viper, gopath string, envs []string) {
	viper.Set("version", "3")
	viper.Set("services.mysqlserver.image", "gruppometasrl/mysql57")
	viper.Set("services.mysqlserver.container_name", "test-mysql")
	viper.Set("services.mysqlserver.volumes", []string{gopath + ":/docker-entrypoint-initdb.d"})
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
