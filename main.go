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
	Name  string
	Envs  []string
	Ports []string
}

func main() {
	var c ConfigDto
	err := Read("", &c)
	if err != nil {
		fmt.Printf("read config error:%v", err)
		return
	}
	if c.IsKafka {
		if err = KafkaCompose(); err != nil {
			fmt.Printf("write to kafka error:%v", err)
			return
		}
	}
	if c.Mysql.IsUpdate {
		if err = MysqlCompose(c.Mysql.Sqlpath, c.Mysql.Envs); err != nil {
			fmt.Printf("write to mysql error:%v", err)
			return
		}
	}
	if err = AppCompose(c.Gopath, c.Project, c.SubProjects); err != nil {
		fmt.Printf("write to app error:%v", err)
		return
	}
	_, err = Cmd("docker-compose -f app-compose.yml up")
	fmt.Println(err)

}

func Cmd(cmd string) (out []byte, err error) {
	if goInfo.GetInfo().GoOS == Windows {
		out, err = exec.Command("cmd", "/c", cmd).Output()
	} else {
		out, err = exec.Command("bash", "-c", cmd).Output()
	}
	if err != nil {
		panic("some error found")
	}
	return
}

func AppCompose(gopath string, project *ProjectDto, subProjects []*ProjectDto) error {
	viper := viper.New()
	appComposeMain(viper, gopath, project)
	for _, project := range subProjects {
		appCompose(viper, gopath, project)
	}
	viper.Set("networks.sharenet.external", true)
	viper.Set("version", "3")
	return viper.WriteConfig()
}

func appComposeMain(viper *viper.Viper, gopath string, project *ProjectDto) {
	servicePre := "services." + project.Name
	viper.SetConfigName("app-compose")
	viper.AddConfigPath(".")

	viper.Set(servicePre+".build.context", gopath+"/src/"+project.Name)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", project.Name)

	viper.Set(servicePre+".container_name", project.Name)
	// viper.Set(servicePre+".working_dir", ".")
	// viper.Set(servicePre+".command", "go test")
	viper.Set(servicePre+".networks", []string{"sharenet"})
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

//env format []string{"MYSQL_ROOT_PASSWORD=1234"}
func appCompose(viper *viper.Viper, gopath string, project *ProjectDto) {
	servicePre := "services." + project.Name
	viper.SetConfigName("app-compose")
	viper.AddConfigPath(".")

	viper.Set(servicePre+".build.context", gopath+"/src/"+project.Name)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", project.Name)

	viper.Set(servicePre+".container_name", project.Name)
	viper.Set(servicePre+".networks", []string{"sharenet"})
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

func MysqlCompose(gopath string, envs []string) error {
	viper := viper.New()
	viper.SetConfigName("mysql-compose")
	viper.AddConfigPath(".")
	viper.Set("version", "3")
	viper.Set("services.mysql.image", "gruppometasrl/mysql57")
	viper.Set("services.mysql.container_name", "local-mysql")
	viper.Set("services.mysql.restart", "always")
	viper.Set("services.mysql.volumes", []string{gopath + ":/docker-entrypoint-initdb.d"})
	viper.Set("services.mysql.ports", []string{"3306:3306"})
	viper.Set("services.mysql.networks", []string{"sharenet"})
	viper.Set("services.mysql.environment", envs)
	viper.Set("networks.sharenet.external", true)
	return viper.WriteConfig()
}

func KafkaCompose() error {
	viper := viper.New()
	viper.SetConfigName("kafka-compose")
	viper.AddConfigPath(".")
	viper.Set("version", "3")
	viper.Set("services.kafkaserver.image", "spotify/kafka:latest")
	viper.Set("services.kafkaserver.container_name", "kafka")
	viper.Set("services.kafkaserver.hostname", "kafkaserver")
	viper.Set("services.kafkaserver.restart", "always")
	viper.Set("services.kafkaserver.networks", []string{"sharenet"})
	viper.Set("services.kafkaserver.ports", []string{"2181:2181", "9092:9092"})
	viper.Set("services.kafkaserver.environment", []string{"ADVERTISED_HOST=kafkaserver",
		"ADVERTISED_PORT=9092"})
	viper.Set("networks.sharenet.external", true)
	return viper.WriteConfig()
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
