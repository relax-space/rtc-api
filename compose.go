package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// generate docker-compose
func setComposeApp(viper *viper.Viper, project *ProjectDto) {
	appComposeMain(viper, project)
	for _, project := range project.SubProjects {
		appComposeSub(viper, project)
	}
	viper.Set("version", "3")
}

func setComposeMysql(viper *viper.Viper, port string) {
	envs := []string{"MYSQL_ROOT_PASSWORD=1234"}
	viper.Set("services.mysqlserver.image", "gruppometasrl/mysql57")
	viper.Set("services.mysqlserver.container_name", "test-mysql")
	viper.Set("services.mysqlserver.volumes", []string{
		".:/docker-entrypoint-initdb.d",
	})
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
		"./redis/redis.conf:/usr/local/etc/redis/redis.conf",
	})
}

func setComposeNginx(viper *viper.Viper, projectName, port string) {
	viper.Set("services.nginx.image", "nginx:latest")
	viper.Set("services.nginx.container_name", "test-nginx")
	viper.Set("services.nginx.ports", []string{port + ":80"})
	viper.Set("services.nginx.restart", "on-failure:20")
	viper.Set("services.nginx.depends_on", []string{projectName + "server"})
	viper.Set("services.nginx.volumes", []string{
		"./nginx/default.conf:/etc/nginx/conf.d/default.conf",
		"./nginx/html:/usr/share/nginx/html",
		"./nginx:/var/log/nginx",
	})

}

func setComposeProducer(viper *viper.Viper, port string, project *ProjectDto) {
	project.ServiceName = EventBroker_Name
	project.Ports = []string{port + ":3000"}
	project.Dependencies = []string{"kafkaserver"}
	appCompose(viper, project, "Dockerfile", "always")

}

func setComposeConsumer(viper *viper.Viper, project *ProjectDto) {
	project.Dependencies = []string{"kafkaserver"}
	dockerfile := fmt.Sprintf("%v/src/pangpang/eventbroker/cmd/kafka-consumer/Dockerfile", os.Getenv("GOPATH"))
	project.Dependencies = []string{"kafkaserver", "mysqlserver", "redisserver"}
	appCompose(viper, project, dockerfile, "always")
}

// utils
func appCompose(viper *viper.Viper, project *ProjectDto, dockerfile, restart string) {
	servicePre := "services." + project.ServiceName + "server"
	viper.Set(servicePre+".build.context", getBuildPath(project.ParentFolderName, project.GitShortPath))
	viper.Set(servicePre+".build.dockerfile", dockerfile)
	viper.Set(servicePre+".image", "test-"+project.ServiceName)
	viper.Set(servicePre+".restart", restart)

	viper.Set(servicePre+".container_name", "test-"+project.ServiceName)
	viper.Set(servicePre+".environment", project.Envs)
	viper.Set(servicePre+".ports", project.Ports)
	viper.Set(servicePre+".depends_on", project.Dependencies)
}

func appComposeMain(viper *viper.Viper, project *ProjectDto) {

	viper.SetConfigName(YmlNameDockerCompose)
	viper.AddConfigPath(TEMP_FILE)

	project.Dependencies = project.SubNames
	for _, subName := range project.SubNames {
		project.Dependencies = append(project.Dependencies, subName+"server")
	}
	appComposeDependency(project)
	appCompose(viper, project, "Dockerfile", "on-failure:10")
}

func getBuildPath(parentFolderName, gitShortPath string) (buildPath string) {
	path := ""
	if len(parentFolderName) != 0 {
		path = "/" + parentFolderName
	}
	lastIndex := strings.LastIndex(gitShortPath, "/")
	pName := gitShortPath[lastIndex+1:]
	buildPath = fmt.Sprintf("%v/src%v/%v", os.Getenv("GOPATH"), path, pName)
	return
}

func appComposeSub(viper *viper.Viper, project *ProjectDto) {
	appComposeDependency(project)
	appCompose(viper, project, "Dockerfile", "on-failure:10")
}

func appComposeDependency(project *ProjectDto) {
	if shouldStartMysql(project) {
		project.Dependencies = append(project.Dependencies, "mysqlserver")
	}
	if shouldStartRedis(project) {
		project.Dependencies = append(project.Dependencies, "redisserver")
	}
	if shouldStartMongo(project) {
		project.Dependencies = append(project.Dependencies, "mongoserver")
	}
	if shouldStartSqlServer(project) {
		project.Dependencies = append(project.Dependencies, "sqlServerserver")
	}
	if shouldStartKakfa(project) {
		project.Dependencies = append(project.Dependencies, "kafkaserver")
	}

}
