package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Compose struct {
	ServiceName string
	ImageName   string
	Restart     string
	Environment []string
	Ports       []string

	DependsOn []string
	Build     struct {
		Context    string
		Dockerfile string
	}
}

// generate docker-compose
func (d Compose) setComposeApp(viper *viper.Viper, project *ProjectDto) {
	d.appComposeMain(viper, project)
	for _, project := range project.SubProjects {
		d.appComposeSub(viper, project)
	}
	viper.Set("version", "3")
}

func (d Compose) setComposeMysql(viper *viper.Viper, port string) {

	serviceName := "mysql"
	servicePre := Compose{}.getServicePre(serviceName)

	viper.Set(servicePre+".image", "mysql:5.7.22")
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".volumes", []string{
		".:/docker-entrypoint-initdb.d",
	})
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Mysql})
	//viper.Set("services.mysqlserver.restart", "always")
	viper.Set(servicePre+".environment", []string{"MYSQL_ROOT_PASSWORD=1234"})
}

func (d Compose) setComposeKafka(viper *viper.Viper, port string) {

	d.setComposeZookeeper(viper, "2181")

	serviceName := "kafka"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", "wurstmeister/kafka")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".hostname", containerName)
	//viper.Set("services.kafkaserver.restart", "always")
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Kafka, outPort.Kafka + ":" + outPort.Kafka})

	viper.Set(servicePre+".environment.KAFKA_LISTENERS", fmt.Sprintf("INSIDE://:%v,OUTSIDE://:%v", inPort.Kafka, outPort.Kafka))
	viper.Set(servicePre+".environment.KAFKA_INTER_BROKER_LISTENER_NAME", "INSIDE")
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_LISTENERS", fmt.Sprintf("INSIDE://test-kafka:%v,OUTSIDE://localhost:%v", inPort.Kafka, outPort.Kafka))
	viper.Set(servicePre+".environment.KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", "INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT")
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", "test-zookeeper:2181")

}

func (d Compose) setComposeZookeeper(viper *viper.Viper, port string) {

	serviceName := "zookeeper"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", "wurstmeister/zookeeper:latest")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + port})
}

func (d Compose) setComposeRedis(viper *viper.Viper, port string) {

	serviceName := "redis"
	servicePre := Compose{}.getServicePre(serviceName)

	viper.Set(servicePre+".image", "redis:latest")
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".hostname", d.getContainerName(serviceName))
	//	viper.Set("services.redisserver.restart", "always")
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Nginx})
	viper.Set(servicePre+".volumes", []string{
		"./redis/redis.conf:/usr/local/etc/redis/redis.conf",
	})
}

func (d Compose) setComposeNginx(viper *viper.Viper, projectName, port string) {

	serviceName := "nginx"
	servicePre := Compose{}.getServicePre(serviceName)

	viper.Set(servicePre+".image", "nginx:latest")
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Nginx})
	//viper.Set("services.nginx.restart", "on-failure:20")
	viper.Set(servicePre+".depends_on", []string{d.getServiceServer(projectName)})
	viper.Set(servicePre+".volumes", []string{
		"./nginx/default.conf:/etc/nginx/conf.d/default.conf",
		"./nginx/html:/usr/share/nginx/html",
		"./nginx:/var/log/nginx",
	})

}

func (d Compose) setComposeProducer(viper *viper.Viper, port string, project *ProjectDto) {
	serviceName := EventBroker_Name
	compose := &Compose{
		ServiceName: serviceName,
		ImageName:   REGISTRYNAME + "/" + serviceName + "-qa",
		//Restart:     "always",
		Environment: project.Envs,
		Ports:       []string{port + ":3000"},
		DependsOn:   []string{d.getWaitName(serviceName)},
	}
	compose.appCompose(viper)
	ComposeWait{}.waitCompose(viper, EventBroker_Name, map[string]NamePortDto{
		"kafka": NamePortDto{Name: "kafka", Port: "9092"},
	})

}

func (d Compose) setComposeConsumer(viper *viper.Viper, project *ProjectDto) {
	compose := &Compose{
		ServiceName: project.ServiceName,
		ImageName:   REGISTRYNAME + "/" + project.ServiceName + "-qa",
		//Restart:     "always",
		Environment: project.Envs,
		Ports:       project.Ports,
		DependsOn:   []string{d.getWaitName(project.ServiceName)},
	}
	compose.appCompose(viper)

	ComposeWait{}.waitCompose(viper, project.ServiceName, map[string]NamePortDto{
		"kafka": NamePortDto{Name: "kafka", Port: inPort.Kafka},
		"mysql": NamePortDto{Name: "mysql", Port: inPort.Mysql},
		"redis": NamePortDto{Name: "redis", Port: inPort.Redis},
	})
}

// utils

func (d Compose) getServicePre(serviceName string) string {
	return "services." + d.getServiceServer(serviceName)
}

func (Compose) getServiceServer(serviceName string) string {
	return serviceName + SUFSERVER
}

func (Compose) getContainerName(serviceName string) string {
	return PRETEST + serviceName
}
func (d *Compose) appCompose(viper *viper.Viper) {
	servicePre := Compose{}.getServicePre(d.ServiceName)

	viper.Set(servicePre+".image", d.ImageName+":latest")
	if len(d.Restart) != 0 {
		viper.Set(servicePre+".restart", d.Restart)
	}
	viper.Set(servicePre+".container_name", d.getContainerName(d.ServiceName))
	viper.Set(servicePre+".environment", d.Environment)
	viper.Set(servicePre+".ports", d.Ports)

	viper.Set(servicePre+".depends_on", d.DependsOn)
}

func (d Compose) appComposeMain(viper *viper.Viper, project *ProjectDto) {

	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)
	compose := &Compose{
		ServiceName: project.ServiceName,
		ImageName:   project.Registry,
		//Restart:     "on-failure:10",
		Environment: project.Envs,
		Ports:       project.Ports,

		DependsOn: []string{d.getWaitName(project.ServiceName)},
	}
	compose.appCompose(viper)
}

func (Compose) getBuildPath(parentFolderName, gitShortPath string) (buildPath string) {
	path := ""
	if len(parentFolderName) != 0 {
		path = "/" + parentFolderName
	}
	lastIndex := strings.LastIndex(gitShortPath, "/")
	pName := gitShortPath[lastIndex+1:]
	buildPath = fmt.Sprintf("%v/src%v/%v", getGoPath(), path, pName)
	return
}

func (d Compose) appComposeSub(viper *viper.Viper, project *ProjectDto) {
	compose := &Compose{
		ServiceName: project.ServiceName,
		ImageName:   project.Registry,
		//Restart:     "on-failure:10",
		Environment: project.Envs,
		Ports:       project.Ports,

		DependsOn: []string{d.getWaitName(project.ServiceName)},
	}
	compose.appCompose(viper)
}
func (Compose) getWaitName(serviceName string) string {
	return PREWAIT + serviceName + SUFSERVER
}
