package main

import (
	"strings"

	"github.com/spf13/viper"
)

type ComposeWait struct {
}

type NamePortDto struct {
	Name string
	Port string
}

func (d ComposeWait) appComposeDependency(project *ProjectDto) (deps map[string]NamePortDto) {

	deps = make(map[string]NamePortDto, 0)

	for _, sub := range project.SubProjects {
		deps[sub.ServiceName] = NamePortDto{
			Name: sub.ServiceName,
			Port: d.getIntranetPort(sub.Ports[0]),
		}
	}

	if shouldStartMysql(project) {
		serviceName := "mysql"
		deps[serviceName] = NamePortDto{
			Name: serviceName,
			Port: inPort.Mysql,
		}
	}
	if shouldStartRedis(project) {
		serviceName := "redis"
		deps[serviceName] = NamePortDto{
			Name: serviceName,
			Port: inPort.Redis,
		}
	}
	if shouldStartMongo(project) {
		serviceName := "mongo"
		deps[serviceName] = NamePortDto{
			Name: serviceName,
			Port: inPort.Mongo,
		}
	}
	if shouldStartSqlServer(project) {
		serviceName := "sqlServer"
		deps[serviceName] = NamePortDto{
			Name: serviceName,
			Port: inPort.SqlServer,
		}
	}
	if shouldStartKakfa(project) {
		serviceName := "kafka"
		deps[serviceName] = NamePortDto{
			Name: serviceName,
			Port: inPort.Kafka,
		}
	}
	return
}

func (d ComposeWait) setWaitCompose(viper *viper.Viper, project *ProjectDto) {
	deps := d.appComposeDependency(project)
	d.waitCompose(viper, project.ServiceName, deps)
	for _, sub := range project.SubProjects {
		subdeps := d.appComposeDependency(sub)
		d.waitCompose(viper, sub.ServiceName, subdeps)
	}
}

func (d ComposeWait) getWaitEnvs(deps map[string]NamePortDto) (envs string) {
	for _, sub := range deps {
		envs += "," + Compose{}.getContainerName(sub.Name) + ":" + sub.Port
	}
	if len(envs) != 0 {
		envs = envs[1:]
	}

	return
}

func (d ComposeWait) waitCompose(viper *viper.Viper, serviceName string, deps map[string]NamePortDto) {

	waitName := Compose{}.getWaitName(serviceName)
	servicePre := "services." + waitName
	envs := d.getWaitEnvs(deps)
	environments := []string{"TARGETS=" + envs, "TIMEOUT=300"}

	depends := make([]string, 0)
	for _, dep := range deps {
		depends = append(depends, Compose{}.getServiceServer(dep.Name))
	}
	viper.Set(servicePre+".image", WAITIMAGE)
	viper.Set(servicePre+".container_name", Compose{}.getContainerName(waitName))
	viper.Set(servicePre+".environment", environments)
	viper.Set(servicePre+".depends_on", depends)
}

func (ComposeWait) getIntranetPort(port string) (newPort string) {
	if strings.Contains(port, ":") {
		newPort = port[strings.Index(port, ":")+1:]
	}
	return
}
