package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	IsComplex      bool   //a git contains multiple microservices
	ServiceName    string //eg. ipay-api
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
		setComposeMysql(viper, c.Mysql.Ports, c.Mysql.Databases)
		setComposeNginx(viper, c.Project.ServiceName)
		setComposeApp(viper, c.Project)

		if err = writeConfig(YmlNameDockerCompose+".yml", viper); err != nil {
			fmt.Printf("write to config.yml error:%v", err)
			return
		}
	}

	//3. run docker-compose
	if shouldRestartData(c.Scope, c.NoCache) {
		if _, err = Cmd("docker-compose", "-f", "docker-compose.yml", "down", "--remove-orphans"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose downed!")
	}

	if shouldRestartApp(c.Scope, c.NoCache) {
		if _, err = Cmd("docker-compose", "-f", "docker-compose.yml", "build"); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
		fmt.Println("==> compose builded!")
	}

	go func() {
		if _, err = Cmd("docker-compose", "-f", "docker-compose.yml", "up"); err != nil {
			fmt.Printf("err:%v", err)
			return
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
