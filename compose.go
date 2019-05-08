package main

import (
	"fmt"
	"strconv"
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

type NamePortDto struct {
	Name string
	Port string
}

// generate docker-compose
func (d Compose) setComposeApp(viper *viper.Viper, project *ProjectDto) {
	viper.Set("version", "3")
	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)
	d.appCompose(viper, project)
	for _, sub := range project.SubProjects {
		d.appCompose(viper, sub)
	}
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

func (d Compose) setComposeKafkaEland(viper *viper.Viper, port, secondPort, zookeeperPort, ip string) {

	portInt, _ := strconv.ParseInt(port, 10, 64)
	jmxPort := 9097

	d.setComposeZookeeperEland(viper, zookeeperPort, ip)
	serviceName := "kafka"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", REGISTRYELAND+"/kafka")
	viper.Set(servicePre+".container_name", containerName)
	//viper.Set(servicePre+".restart", "always")
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Kafka, fmt.Sprintf("%v:%v", jmxPort, jmxPort)})

	viper.Set(servicePre+".environment.KAFKA_BROKER_ID", 1)
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_HOST_NAME", ip)
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_PORT", portInt)
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", "zoo1:"+inPort.Zookeeper)
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECTION_TIMEOUT_MS", 60000)

	viper.Set(servicePre+".environment.KAFKA_DELETE_TOPIC_ENABLE", "true")
	viper.Set(servicePre+".environment.KAFKA_LOG_DIRS", "/kafka/kafka-logs-24bf1bde016a")
	viper.Set(servicePre+".environment.KAFKA_LOG_RETENTION_HOURS", 120)
	viper.Set(servicePre+".environment.KAFKA_LOG_CLEANUP_POLICY", "delete")
	viper.Set(servicePre+".environment.KAFKA_LOG_CLEANER_ENABLE", "true")

	viper.Set(servicePre+".environment.KAFKA_JVM_PERFORMANCE_OPTS", "-XX:+UseG1GC -XX:MaxGCPauseMillis=20 -XX:InitiatingHeapOccupancyPercent=35 -XX:+DisableExplicitGC -Djava.awt.headless=true")
	viper.Set(servicePre+".environment.KAFKA_HEAP_OPTS", "-Xmx1G")
	viper.Set(servicePre+".environment.JMX_PORT", jmxPort)
	viper.Set(servicePre+".environment.KAFKA_JMX_OPTS",
		fmt.Sprintf("-Dcom.sun.management.jmxremote=true -Dcom.sun.management.jmxremote.authenticate=false  -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.rmi.port=%v -Djava.rmi.server.hostname=%v", jmxPort, ip))

	viper.Set(servicePre+".extra_hosts", []string{fmt.Sprintf("zoo1:%v", ip)})

}

func (d Compose) setComposeZookeeperEland(viper *viper.Viper, port, ip string) {

	serviceName := "zookeeper"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", REGISTRYELAND+"/zookeeper")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Zookeeper, "2888:2888", "3888:3888"})
	viper.Set(servicePre+".environment.ZOO_MY_ID", 1)
	viper.Set(servicePre+".environment.ZOO_SERVERS", "server.1=0.0.0.0:2888:3888")
	viper.Set(servicePre+".extra_hosts", []string{fmt.Sprintf("zoo1:%v", ip)})
}

func (d Compose) setComposeKafka(viper *viper.Viper, port, secondPort, zookeeperPort, ip string) {

	d.setComposeZookeeper(viper, zookeeperPort)

	serviceName := "kafka"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", "wurstmeister/kafka")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".hostname", containerName)
	//viper.Set("services.kafkaserver.restart", "always")
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Kafka, secondPort + ":" + inPort.KafkaSecond})

	viper.Set(servicePre+".environment.KAFKA_LISTENERS", fmt.Sprintf("INSIDE://:%v,OUTSIDE://:%v", inPort.Kafka, secondPort))
	viper.Set(servicePre+".environment.KAFKA_INTER_BROKER_LISTENER_NAME", "INSIDE")
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_LISTENERS",
		fmt.Sprintf("INSIDE://%v:%v,OUTSIDE://localhost:%v", containerName, inPort.Kafka, secondPort))
	viper.Set(servicePre+".environment.KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", "INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT")
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", d.getContainerName("zookeeper")+":"+inPort.Zookeeper)

}

func (d Compose) setComposeZookeeper(viper *viper.Viper, port string) {

	serviceName := "zookeeper"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", "wurstmeister/zookeeper:latest")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Zookeeper})
}

func (d Compose) setComposeRedis(viper *viper.Viper, port string) {

	serviceName := "redis"
	servicePre := Compose{}.getServicePre(serviceName)

	viper.Set(servicePre+".image", "redis:3.2.11")
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

	viper.Set(servicePre+".image", "nginx:1.16")
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
		ImageName:   REGISTRYELAND + "/" + serviceName + "-" + app_env,
		//Restart:     "always",
		Environment: project.Envs,
		Ports:       []string{port + ":" + inPort.EventBroker},
		DependsOn:   []string{d.getServiceServer("kafka")},
	}
	compose.setCompose(viper)
}

func (d Compose) setComposeConsumer(viper *viper.Viper, project *ProjectDto) {
	compose := &Compose{
		ServiceName: project.ServiceName,
		ImageName:   REGISTRYELAND + "/" + project.ServiceName + "-" + app_env,
		//Restart:     "always",
		Environment: project.Envs,
		Ports:       project.Ports,
		DependsOn: []string{d.getServiceServer("kafka"),
			d.getServiceServer("mysql"), d.getServiceServer("redis")},
	}
	compose.setCompose(viper)
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
func (d *Compose) setCompose(viper *viper.Viper) {
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

func (d Compose) appCompose(viper *viper.Viper, project *ProjectDto) {

	deps := d.dependency(project)
	compose := &Compose{
		ServiceName: project.ServiceName,
		ImageName:   project.Registry,
		//Restart:     "on-failure:10",
		Environment: project.Envs,
		Ports:       project.Ports,

		DependsOn: deps,
	}
	compose.setCompose(viper)
}

func (d Compose) dependency(project *ProjectDto) (depends []string) {
	deps := d.setComposeDependency(project)
	depends = make([]string, 0)
	for _, dep := range deps {
		depends = append(depends, d.getServiceServer(dep.Name))
	}
	return
}

func (d Compose) setComposeDependency(project *ProjectDto) (deps map[string]NamePortDto) {

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

func (Compose) getIntranetPort(port string) (newPort string) {
	if strings.Contains(port, ":") {
		newPort = port[strings.Index(port, ":")+1:]
	}
	return
}
