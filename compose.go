package main

import (
	"os"

	"github.com/spf13/viper"
)

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
