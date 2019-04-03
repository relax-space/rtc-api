package main

import (
	"os"
	"strings"

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

	lastIndex := strings.LastIndex(project.GitShortPath, "/")
	pName := project.GitShortPath[lastIndex:]

	servicePre := "services." + project.ServiceName
	viper.SetConfigName(YmlNameDockerCompose)
	viper.AddConfigPath(".")

	project.SubNames = append(project.SubNames, "kafkaserver")
	project.SubNames = append(project.SubNames, "mysqlserver")
	viper.Set(servicePre+".build.context", os.Getenv("GOPATH")+"/src/"+pName)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", "test-"+project.ServiceName)
	viper.Set(servicePre+".restart", "on-failure:5")

	viper.Set(servicePre+".container_name", "test-"+project.ServiceName)
	viper.Set(servicePre+".depends_on", project.SubNames)
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

//env format []string{"MYSQL_ROOT_PASSWORD=1234"}
func appCompose(viper *viper.Viper, project *ProjectDto) {
	lastIndex := strings.LastIndex(project.GitShortPath, "/")
	pName := project.GitShortPath[lastIndex:]

	servicePre := "services." + project.ServiceName
	viper.Set(servicePre+".build.context", os.Getenv("GOPATH")+"/src/"+pName)
	viper.Set(servicePre+".build.dockerfile", "Dockerfile")
	viper.Set(servicePre+".image", "test-"+project.ServiceName)
	viper.Set(servicePre+".restart", "on-failure:5")

	viper.Set(servicePre+".depends_on", []string{"mysqlserver"})
	viper.Set(servicePre+".container_name", "test-"+project.ServiceName)
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".environment", project.Envs)
}

func setComposeMysql(viper *viper.Viper, port string) {
	envs := []string{"MYSQL_ROOT_PASSWORD=1234"}
	viper.Set("services.mysqlserver.image", "gruppometasrl/mysql57")
	viper.Set("services.mysqlserver.container_name", "test-mysql")
	viper.Set("services.mysqlserver.volumes", []string{".:/docker-entrypoint-initdb.d"})
	viper.Set("services.mysqlserver.ports", []string{port + ":3306"})
	viper.Set("services.mysqlserver.restart", "always")
	viper.Set("services.mysqlserver.environment", envs)
}

func setComposeKafka(viper *viper.Viper, port string) {
	viper.Set("services.kafkaserver.image", "spotify/kafka:latest")
	viper.Set("services.kafkaserver.container_name", "test-kafka")
	viper.Set("services.kafkaserver.hostname", "kafkaserver")
	viper.Set("services.kafkaserver.restart", "always")
	viper.Set("services.kafkaserver.ports", []string{port + ":9092"})
	viper.Set("services.kafkaserver.environment", []string{"ADVERTISED_HOST=kafkaserver",
		"ADVERTISED_PORT=9092"})
}

func setComposeRedis(viper *viper.Viper, port string) {
	viper.Set("services.redisserver.image", "redis:latest")
	viper.Set("services.redisserver.container_name", "test-redis")
	viper.Set("services.redisserver.hostname", "redisserver")
	viper.Set("services.redisserver.restart", "always")
	viper.Set("services.redisserver.ports", []string{port + ":6379"})
	viper.Set("services.redisserver.volumes", []string{
		"./redis.conf:/usr/local/etc/redis/redis.conf",
	})
}

func setComposeNginx(viper *viper.Viper, projectName string) {
	viper.Set("services.nginx.image", "nginx:latest")
	viper.Set("services.nginx.container_name", "test-nginx")
	viper.Set("services.nginx.ports", []string{"3001:80"})
	viper.Set("services.nginx.restart", "on-failure:5")
	viper.Set("services.nginx.depends_on", []string{projectName})
	viper.Set("services.nginx.volumes", []string{
		"./default.conf:/etc/nginx/conf.d/default.conf",
		"./html:/usr/share/nginx/html",
		".:/var/log/nginx",
	})

}
