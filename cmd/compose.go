package cmd

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

func (d Compose) Exec(c *FullDto, flag *Flag, ip string) (err error) {

	if BoolPointCheck(flag.NoLogin) == false {
		Info("==> docker login " + comboResource.Registry + " ...")
		if _, err = CmdRealtime("docker", "login", "-u", "eland", "-p", registryPwd, comboResource.Registry); err != nil {
			return
		}
	}

	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
		return
	}
	Info("==> compose downed!")
	if BoolPointCheck(flag.NoPull) == false {
		if err = d.checkLatest(dockercompose, c); err != nil {
			return
		}
	}

	if (File{}).WriteString(TEMP_FILE+"/wait-for.sh", wait_for); err != nil {
		return err
	}
	project := *(c.Project)
	if err = d.checkAll(project, c.Port, ip, dockercompose); err != nil {
		return
	}
	Info("check is ok.")

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return
	}
	Info(`==> compose up!`)
	return
}

func (d Compose) checkLatest(dockercompose string, c *FullDto) (err error) {
	if scope == LOCAL.String() {
		return
	}
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "pull"); err != nil {
		return
	}
	Info("==> compose pulled!")

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "build"); err != nil {
		return
	}
	Info("==> compose builded!")
	return
}

func (d Compose) setComposeApp(viper *viper.Viper, project *ProjectDto) {
	viper.Set("version", "3")
	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)
	project.DependsOn = d.dependency(project)
	d.appCompose(viper, project)

	for _, sub := range project.SubProjects {
		sub.DependsOn = d.dependency(sub)
		d.appCompose(viper, sub)
	}
}

func (d Compose) getRegistryCommon(name string, isDefault bool) string {
	registry := name
	if isDefault {
		registry = comboResource.Registry + "/" + name
	}
	if len(comboResource.RegistryCommon) != 0 {
		registry = comboResource.RegistryCommon + "/" + name
	}
	return registry
}

func (d Compose) setComposeMysql(viper *viper.Viper, port string) {

	serviceName := "mysql"
	servicePre := Compose{}.getServicePre(serviceName)
	viper.Set(servicePre+".image", d.getRegistryCommon("mysql:5.7.22", false))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".volumes", []string{
		"./database/mysql/:/docker-entrypoint-initdb.d",
	})
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Mysql})
	viper.Set(servicePre+".environment", []string{"MYSQL_ROOT_PASSWORD=1234"})
}

func (d Compose) setComposeSqlserver(viper *viper.Viper, port string) {

	serviceName := "sqlserver"
	servicePre := Compose{}.getServicePre(serviceName)
	//docker-hub: genschsa/mssql-server-linux
	viper.Set(servicePre+".image", d.getRegistryCommon("mssql-server-linux", true))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.SqlServer})
	viper.Set(servicePre+".volumes", []string{
		"./database/sqlserver/:/docker-entrypoint-initdb.d",
	})
	//MSSQL_PID=Developer,Express,Standard,Enterprise,EnterpriseCore SA_PASSWORD
	viper.Set(servicePre+".environment", []string{"ACCEPT_EULA=Y", "MSSQL_SA_PASSWORD=Eland123", "MSSQL_PID=Developer"})
}

func (d Compose) setComposeKafkaEland(viper *viper.Viper, port, secondPort, zookeeperPort, ip string) {

	portInt, _ := strconv.ParseInt(port, 10, 64)
	jmxPort := 9097

	d.setComposeZookeeperEland(viper, zookeeperPort)
	serviceName := "kafka"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)
	hostName := containerName

	viper.Set(servicePre+".image", d.getRegistryCommon("kafka", true))
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Kafka, fmt.Sprintf("%v:%v", jmxPort, jmxPort)})

	viper.Set(servicePre+".environment.KAFKA_BROKER_ID", 1)
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_HOST_NAME", hostName)
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_PORT", portInt)
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", d.getContainerName("zookeeper")+":"+inPort.Zookeeper)
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
		fmt.Sprintf("-Dcom.sun.management.jmxremote=true -Dcom.sun.management.jmxremote.authenticate=false  -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.rmi.port=%v -Djava.rmi.server.hostname=%v", jmxPort, hostName))

	viper.Set(servicePre+".extra_hosts", []string{fmt.Sprintf("%v:%v", hostName, ip)})

}

func (d Compose) setComposeZookeeperEland(viper *viper.Viper, port string) {

	serviceName := "zookeeper"
	servicePre := Compose{}.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", d.getRegistryCommon("zookeeper", true))
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Zookeeper, "2888:2888", "3888:3888"})
	viper.Set(servicePre+".environment.ZOO_MY_ID", 1)
	viper.Set(servicePre+".environment.ZOO_SERVERS", fmt.Sprintf("server.1=%v:2888:3888", containerName))
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
		fmt.Sprintf("INSIDE://%v:%v,OUTSIDE://%v:%v", containerName, inPort.Kafka, ip, secondPort))
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

	viper.Set(servicePre+".image", d.getRegistryCommon("redis:3.2.11", false))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".hostname", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Redis})
}

func (d Compose) setComposeNginx(viper *viper.Viper, project *ProjectDto, port string) {

	names, _ := d.waitCommandAllProject(project)
	serviceName := "nginx"
	servicePre := Compose{}.getServicePre(serviceName)

	viper.Set(servicePre+".image", d.getRegistryCommon("nginx:1.16", false))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + inPort.Nginx})
	viper.Set(servicePre+".restart", "always")
	viper.Set(servicePre+".depends_on", names)
	viper.Set(servicePre+".volumes", []string{
		"./nginx/default.conf:/etc/nginx/conf.d/default.conf",
		"./nginx/html:/usr/share/nginx/html",
	})

}

func (d Compose) setComposeWaitStart(viper *viper.Viper, project *ProjectDto) {

	names, namePorts := d.waitCommandAllProject(project)
	names = append(names, d.getServiceServer("nginx"))
	namePorts = append(namePorts,
		d.getContainerName("nginx")+":"+inPort.Nginx)

	serviceName := "wait-service"
	servicePre := Compose{}.getServicePre(serviceName)
	viper.Set(servicePre+".image", d.getRegistryCommon("alpine", false))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".depends_on", names)
	viper.Set(servicePre+".volumes", []string{"./wait-for.sh:/go/bin/wait-for.sh"})
	command := fmt.Sprintf("sh -c '/go/bin/wait-for.sh %v -t 36000 -- echo \"wait-service is up!\" && sleep 100h'",
		strings.Join(namePorts, " "))
	viper.Set(servicePre+".command", command)

}
func (d Compose) waitCommandAllProject(project *ProjectDto) ([]string, []string) {
	names, namePorts := d.getAllProjectAndPort(project.SubProjects)
	names = append(names, d.getServiceServer(project.ServiceName))
	namePorts = append(namePorts, d.getContainerName(project.ServiceName)+":"+project.Ports[0])
	return names, namePorts
}

func (d Compose) getAllProjectAndPort(projectList []*ProjectDto) ([]string, []string) {
	names := make([]string, 0)
	namePorts := make([]string, 0)
	for _, p := range projectList {
		names = append(names, d.getServiceServer(p.ServiceName))
		namePort := d.getContainerName(p.ServiceName) + ":" + p.Ports[0]
		namePorts = append(namePorts, namePort)
		if p.SubProjects != nil && len(p.SubProjects) != 0 {
			d.getAllProjectAndPort(p.SubProjects)
		}
	}
	return names, namePorts
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
func (d Compose) appCompose(viper *viper.Viper, project *ProjectDto) {
	servicePre := Compose{}.getServicePre(project.ServiceName)

	viper.Set(servicePre+".image", project.Registry+":latest")

	viper.Set(servicePre+".container_name", d.getContainerName(project.ServiceName))
	viper.Set(servicePre+".environment", project.Envs)
	viper.Set(servicePre+".ports", project.Ports)

	viper.Set(servicePre+".depends_on", project.DependsOn)
	viper.Set(servicePre+".volumes", []string{"./wait-for.sh:/go/bin/wait-for.sh"})
	viper.Set(servicePre+".command", d.setAppCommand(project.Entrypoint, project.DependsOn))

}

func (d Compose) dependency(project *ProjectDto) []string {
	deps := make(map[string]string, 0)
	d.subDependency(project.SubProjects, deps)
	if (ProjectInfo{}).ShouldKafka(project) {
		deps["kafka"] = ""
	}
	list := Database{}.All(project, true)
	for k := range list {
		deps[k] = ""
	}
	depends := make([]string, 0)
	for k := range deps {
		depends = append(depends, d.getServiceServer(k))
	}
	return depends
}

func (d Compose) subDependency(projectList []*ProjectDto, deps map[string]string) {
	for _, sub := range projectList {
		deps[strings.ToLower(sub.ServiceName)] = ""
		if sub.SubProjects != nil && len(sub.SubProjects) != 0 {
			d.subDependency(sub.SubProjects, deps)
		}
	}
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

func dockerWaitCommand(entrypoint string, ipPorts []IpPortDto) string {
	var nameports string

	for _, ipPort := range ipPorts {
		nameports += ipPort.Ip + ":" + ipPort.Port + " "
	}
	command := fmt.Sprintf("sh -c '/go/bin/wait-for.sh %v -t 36000 -- %v'",
		nameports, entrypoint)
	return command

}

func (d Compose) setAppCommand(entrypoint string, serviceNames []string) string {
	newIpPorts := make([]IpPortDto, 0)
	databases := []IpPortDto{
		IpPortDto{Ip: d.getContainerName("kafka"), Port: inPort.Kafka},
		IpPortDto{Ip: d.getContainerName("mysql"), Port: inPort.Mysql},
		IpPortDto{Ip: d.getContainerName("sqlserver"), Port: inPort.SqlServer},
	}
	for _, name := range serviceNames {
		name := PRETEST + strings.TrimSuffix(name, SUFSERVER)
		for _, dto := range databases {
			if name == dto.Ip {
				newIpPorts = append(newIpPorts, dto)
			}
		}
	}
	return dockerWaitCommand(entrypoint, newIpPorts)
}

func (d Compose) checkAll(project ProjectDto, port PortDto, ip, dockercompose string) (err error) {

	dt := Database{}
	if dt.ShouldDbLoop(&project, MYSQL) {
		if err = d.checkMysql(dockercompose, port.Mysql, ip); err != nil {
			return
		}
	}
	if dt.ShouldDbLoop(&project, SQLSERVER) {
		if err = d.checkSqlServer(dockercompose, port.SqlServer, ip); err != nil {
			return
		}
	}
	if (ProjectInfo{}).ShouldKafka(&project) {
		if err = d.checkKafka(dockercompose, port.Kafka, ip); err != nil {
			return
		}
	}
	return
}

func (d Compose) checkMysql(dockercompose, port, ip string) (err error) {
	dbType := MYSQL.String()

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", dbType+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	Info(fmt.Sprintf("begin ping %v,%v:%v", dbType, ip, port))
	db, err := sql.Open("mysql", fmt.Sprintf("root:1234@tcp(%v:%v)/mysql?charset=utf8", ip, port))
	if err != nil {
		return
	}
	defer db.Close()
	//remove mysql log
	buffer := bytes.NewBuffer(make([]byte, 0, 64))
	logger := log.New(buffer, "prefix: ", 0)
	mysql.SetLogger(logger)

	for index := 1; index < 300; index++ {
		err = db.Ping()
		if err != nil {
			time.Sleep(2 * time.Second)
			if index%30 == 0 {
				Info(err.Error())
			}
			continue
		}
		err = nil
		break
	}
	if err != nil {
		return
	}
	Info("finish ping " + dbType)
	return
}

func (d Compose) checkSqlServer(dockercompose, port, ip string) (err error) {

	dbType := SQLSERVER.String()

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", dbType+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	Info(fmt.Sprintf("begin ping %v,%v:%v", dbType, ip, port))
	db, err := sql.Open("sqlserver",
		fmt.Sprintf("sqlserver://sa:Eland123@%v:%v?database=master", ip, port))

	if err != nil {
		return
	}
	defer db.Close()
	for index := 1; index < 300; index++ {
		err = db.Ping()
		if err != nil {
			time.Sleep(2 * time.Second)
			if index%30 == 0 {
				Info(err.Error())
			}
			continue
		}
		err = nil
		break
	}
	if err != nil {
		return
	}
	Info("finish ping " + dbType)
	return
}

func (d Compose) checkKafka(dockercompose, port, ip string) (err error) {
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "zookeeper"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "kafka"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	Info(fmt.Sprintf("begin ping kafka%v:%v", ip, port))
	for index := 1; index < 300; index++ {
		if err = d.dailKafka(port, ip); err != nil {
			time.Sleep(2 * time.Second)
			if index%30 == 0 {
				Info(err.Error())
			}
			continue
		}
		err = nil
		break
	}
	if err != nil {
		return
	}
	Info("finish ping kafka")
	return
}

func (d Compose) dailKafka(port, ip string) (err error) {
	_, err = kafkautil.DialLeader(context.Background(), "tcp", ip+":"+port, "ping", 0)
	if err != nil {
		return
	}
	return
}
