package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	kafkautil "github.com/segmentio/kafka-go"

	mysql "github.com/go-sql-driver/mysql"
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

func (d Compose) WriteYml(viper *viper.Viper) (err error) {

	ymlStr, err := getStringViper(viper)
	if err != nil {
		err = fmt.Errorf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		return
	}

	ymlStr = d.upperKafkaEnvEland(ymlStr)

	if (File{}).WriteString(TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", ymlStr); err != nil {
		err = fmt.Errorf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
		return
	}
	return
}

func (d Compose) Exec(c *FullDto) (err error) {

	fmt.Println("==> docker login " + REGISTRYELAND + " ...")

	if _, err = CmdRealtime("docker", "login", "-u", "eland", "-p", registryPwd, REGISTRYELAND); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
		return
	}
	fmt.Println("==> compose downed!")
	if err = d.checkLatest(dockercompose, c); err != nil {
		return
	}
	project := *(c.Project)
	if err = d.checkAll(project, c.Port, dockercompose); err != nil {
		return
	}
	fmt.Println("check is ok.")
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return
	}
	fmt.Printf(`==> compose up! you can start testing now.view status:docker ps -a.`)
	return
}

func (d Compose) checkLatest(dockercompose string, c *FullDto) (err error) {
	if scope == LOCAL.String() {
		return
	}
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "pull"); err != nil {
		return
	}
	fmt.Println("==> compose pulled!")

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "build"); err != nil {
		return
	}
	fmt.Println("==> compose builded!")
	return
}

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
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_HOST_NAME", "test-kafka")
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_PORT", portInt)
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", "test-kafka:"+inPort.Zookeeper)
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

	viper.Set(servicePre+".extra_hosts", []string{fmt.Sprintf("test-kafka:%v", ip)})

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
	viper.Set(servicePre+".extra_hosts", []string{fmt.Sprintf("test-kafka:%v", ip)})
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
		fmt.Sprintf("INSIDE://%v:%v,OUTSIDE://127.0.0.1:%v", containerName, inPort.Kafka, secondPort))
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
	viper.Set(servicePre+".restart", "on-failure:10")
	viper.Set(servicePre+".depends_on", []string{d.getServiceServer(projectName)})
	viper.Set(servicePre+".volumes", []string{
		"./nginx/default.conf:/etc/nginx/conf.d/default.conf",
		"./nginx/html:/usr/share/nginx/html",
		//"./nginx:/var/log/nginx",
	})

}

func (d Compose) getServicePre(serviceName string) string {
	return "services." + d.getServiceServer(serviceName)
}

func (Compose) getServiceServer(serviceName string) string {
	return strings.ToLower(serviceName) + SUFSERVER
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

func (d Compose) appCompose(viper *viper.Viper, project *ProjectDto) {

	deps := d.dependency(project)
	compose := &Compose{
		ServiceName: project.ServiceName,
		ImageName:   project.Registry,
		Restart:     "on-failure:10",
		Environment: project.Envs,
		Ports:       project.Ports,

		DependsOn: deps,
	}
	compose.setCompose(viper)
}

func (d Compose) dependency(project *ProjectDto) (depends []string) {
	deps := d.setComposeDependency(project)
	depends = make([]string, 0)
	for dep := range deps {
		depends = append(depends, d.getServiceServer(dep))
	}
	return
}

func (d Compose) setComposeDependency(project *ProjectDto) (deps map[string]string) {

	deps = make(map[string]string, 0)

	for _, sub := range project.SubProjects {
		deps[strings.ToLower(sub.ServiceName)] = ""
	}
	p := ProjectInfo{}
	if p.ShouldDb(project, MYSQL) {
		deps[MYSQL.String()] = ""
	}
	if p.ShouldDb(project, REDIS) {
		deps[REDIS.String()] = ""
	}
	if p.ShouldDb(project, MONGO) {
		deps[MONGO.String()] = ""
	}
	if p.ShouldDb(project, SQLSERVER) {
		deps[SQLSERVER.String()] = ""
	}

	if p.ShouldKafka(project) {
		deps["kafka"] = ""
	}
	return
}

func (d Compose) upperKafkaEnvEland(ymlStr string) string {
	ymlStr = strings.Replace(ymlStr, "kafka_broker_id", "KAFKA_BROKER_ID", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_host_name", "KAFKA_ADVERTISED_HOST_NAME", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_port", "KAFKA_ADVERTISED_PORT", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connection_timeout_ms", "KAFKA_ZOOKEEPER_CONNECTION_TIMEOUT_MS", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_delete_topic_enable", "KAFKA_DELETE_TOPIC_ENABLE", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_dirs", "KAFKA_LOG_DIRS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_retention_hours", "KAFKA_LOG_RETENTION_HOURS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_cleanup_policy", "KAFKA_LOG_CLEANUP_POLICY", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_log_cleaner_enable", "KAFKA_LOG_CLEANER_ENABLE", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_jvm_performance_opts", "KAFKA_JVM_PERFORMANCE_OPTS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_heap_opts", "KAFKA_HEAP_OPTS", -1)
	ymlStr = strings.Replace(ymlStr, "jmx_port", "JMX_PORT", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_jmx_opts", "KAFKA_JMX_OPTS", -1)

	ymlStr = strings.Replace(ymlStr, "zoo_my_id", "ZOO_MY_ID", -1)
	ymlStr = strings.Replace(ymlStr, "zoo_servers", "ZOO_SERVERS", -1)

	ymlStr = d.upperKafkaEnv(ymlStr)

	return ymlStr
}

func (d Compose) upperKafkaEnv(ymlStr string) string {
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_listeners", "KAFKA_ADVERTISED_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_inter_broker_listener_name", "KAFKA_INTER_BROKER_LISTENER_NAME", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listener_security_protocol_map", "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listeners", "KAFKA_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_advertised_port", "KAFKA_ADVERTISED_PORT", -1)
	return ymlStr
}

func (d Compose) checkAll(project ProjectDto, port PortDto, dockercompose string) (err error) {

	p := ProjectInfo{}
	if p.ShouldDb(&project, MYSQL) {
		if err = d.checkMysql(dockercompose, port.Mysql); err != nil {
			return
		}
	}
	if p.ShouldKafka(&project) {
		if err = d.checkKafka(dockercompose, port.Kafka); err != nil {
			return
		}
	}

	return
}

func (d Compose) checkMysql(dockercompose, port string) (err error) {

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "mysql"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:1234@tcp(127.0.0.1:%v)/mysql?charset=utf8", port))
	if err != nil {
		fmt.Println("mysql", err)
		return
	}
	//remove mysql log
	buffer := bytes.NewBuffer(make([]byte, 0, 64))
	logger := log.New(buffer, "prefix: ", 0)
	mysql.SetLogger(logger)

	fmt.Println("begin ping db")
	for index := 0; index < 300; index++ {
		err = db.Ping()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		err = nil
		break
	}
	if err != nil {
		fmt.Println("error ping db")
		return
	}
	fmt.Println("finish ping db")
	return
}

func (d Compose) checkKafka(dockercompose, port string) (err error) {
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "zookeeper"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "kafka"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	fmt.Println("begin ping kafka,127.0.0.1:" + port)
	for index := 0; index < 300; index++ {
		if err = d.dailKafka(port); err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		err = nil
		break
	}
	if err != nil {
		fmt.Println("error ping kafka")
		return
	}
	fmt.Println("finish ping kafka")
	return
}

func (d Compose) dailKafka(port string) (err error) {
	_, err = kafkautil.DialLeader(context.Background(), "tcp", "127.0.0.1:"+port, "ping", 0)
	if err != nil {
		return
	}
	return
}
