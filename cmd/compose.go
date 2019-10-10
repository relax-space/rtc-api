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
	InPort PortDto
}

func (d *Compose) SetPort() {
	d.InPort = PortDto{
		Mysql:     "3306",
		Redis:     "6379",
		Mongo:     "27017",
		SqlServer: "1433",
		Kafka:     "9092",

		KafkaSecond: "29092",
		//EventBroker: "3000",
		Nginx:     "80",
		Zookeeper: "2181",
	}
}

func (d *Compose) Write(project *Project, flag *Flag) error {
	d.SetPort()
	viper := viper.New()
	viper.Set("version", "3")
	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)
	if project.Owner.IsKafka {
		ip := *flag.HostIp
		if err := CheckHost(ip); err != nil {
			return err
		}
		d.setKafkaEland(viper, *flag.KafkaPort, ip, flag.RegistryCommon)
	}
	if project.Owner.IsMysql {
		d.setMysql(viper, *flag.MysqlPort, flag.RegistryCommon)
	}
	if project.Owner.IsRedis {
		d.setRedis(viper, *flag.RedisPort, flag.RegistryCommon)
	}
	if project.Owner.IsSqlServer {
		d.setSqlServer(viper, *flag.SqlServerPort, flag.RegistryCommon)
	}
	if project.Owner.IsStream {
		d.setEventBroker(viper, project, *flag.ImageEnv, *flag.HostIp)
	}
	d.setNginx(viper, project, *flag.NginxPort, flag.RegistryCommon)
	d.SetApp(viper, project, *flag.ImageEnv)
	if err := d.WriteYml(viper); err != nil {
		return err
	}
	return nil
}
func (d *Compose) WriteYml(viper *viper.Viper) error {
	ymlStr, err := getStringViper(viper)
	if err != nil {
		return fmt.Errorf("write to %v error:%v", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", err)
	}

	ymlStr = d.upperKafkaEnvEland(ymlStr)

	if (File{}).WriteString("", TEMP_FILE+"/"+YMLNAMEDOCKERCOMPOSE+".yml", ymlStr); err != nil {
	}
	return nil
}

func (d *Compose) Exec(project *Project, flag *Flag) error {
	if BoolPointCheck(flag.NoLogin) == false {
		for _, r := range project.Owner.ImageAccounts {
			Info("==> docker login " + r.Registry + " ...")
			if _, err := CmdRealtime("docker", "login", "-u", r.LoginName, "-p", r.Pwd, r.Registry); err != nil {
				return err
			}
		}
	}

	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
		return err
	}
	Info("==> compose downed!")
	if BoolPointCheck(flag.NoPull) == false {
		if err := d.checkLatest(dockercompose); err != nil {
			return err
		}
	}

	if err := d.checkAll(project, flag, dockercompose); err != nil {
		return err
	}
	Info("check is ok.")

	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return err
	}
	Info(`==> compose up!`)
	return nil
}

func (d *Compose) checkLatest(dockercompose string) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "pull"); err != nil {
		return err
	}
	Info("==> compose pulled!")

	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "build"); err != nil {
		return err
	}
	Info("==> compose builded!")
	return nil
}

func (d *Compose) SetApp(viper *viper.Viper, project *Project, imageEnv string) {
	d.SetAppLoop(viper, []*Project{project}, imageEnv)
}

func (d *Compose) SetAppLoop(viper *viper.Viper, projects []*Project, imageEnv string) {
	for _, project := range projects {
		project.DependsOn = d.dependsOn(project.DependsOn)
		project.Setting.Image = project.Setting.Image + "-" + imageEnv
		d.appCompose(viper, project)
		if len(project.Children) > 0 {
			d.SetAppLoop(viper, project.Children, imageEnv)
		}
	}
}

func (d *Compose) GetImage(name string, registryCommon *string) string {
	var registry string
	if name == "kafka" || name == "mssql-server-linux" || name == "zookeeper" {
		registry = REGISTRYCOMMON
	}
	if StringPointCheck(registryCommon) {
		registry = *registryCommon
	}
	if len(registry) != 0 {
		name = registry + "/" + name
	}
	return name
}

func (d *Compose) setMysql(viper *viper.Viper, port string, registryCommon *string) {

	serviceName := "mysql"
	servicePre := d.getServicePre(serviceName)
	viper.Set(servicePre+".image", d.GetImage("mysql:5.7.22", registryCommon))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".volumes", []string{
		"./database/mysql/:/docker-entrypoint-initdb.d",
	})
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.Mysql})
	viper.Set(servicePre+".environment", []string{"MYSQL_ROOT_PASSWORD=1234"})
}

func (d *Compose) setSqlServer(viper *viper.Viper, port string, registryCommon *string) {

	serviceName := "sqlserver"
	servicePre := d.getServicePre(serviceName)
	//docker-hub: genschsa/mssql-server-linux
	viper.Set(servicePre+".image", d.GetImage("mssql-server-linux", registryCommon))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.SqlServer})
	viper.Set(servicePre+".volumes", []string{
		"./database/sqlserver/:/docker-entrypoint-initdb.d",
	})
	//MSSQL_PID=Developer,Express,Standard,Enterprise,EnterpriseCore SA_PASSWORD
	viper.Set(servicePre+".environment", []string{"ACCEPT_EULA=Y", "MSSQL_SA_PASSWORD=Eland123", "MSSQL_PID=Developer"})
}

func (d *Compose) setKafkaEland(viper *viper.Viper, port, ip string, registryCommon *string) {
	portInt, _ := strconv.ParseInt(port, 10, 64)
	jmxPort := 9097

	d.setZookeeperEland(viper, registryCommon)
	serviceName := "kafka"
	servicePre := d.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)
	hostName := containerName

	viper.Set(servicePre+".image", d.GetImage("kafka", registryCommon))
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.Kafka, fmt.Sprintf("%v:%v", jmxPort, jmxPort)})

	viper.Set(servicePre+".environment.KAFKA_BROKER_ID", 1)
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_HOST_NAME", hostName)
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_PORT", portInt)
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", d.getContainerName("zookeeper")+":"+d.InPort.Zookeeper)
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

func (d *Compose) setZookeeperEland(viper *viper.Viper, registryCommon *string) {

	serviceName := "zookeeper"
	servicePre := d.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", d.GetImage("zookeeper", registryCommon))
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{d.InPort.Zookeeper, "2888:2888", "3888:3888"})
	viper.Set(servicePre+".environment.ZOO_MY_ID", 1)
	viper.Set(servicePre+".environment.ZOO_SERVERS", fmt.Sprintf("server.1=%v:2888:3888", containerName))
}

func (d *Compose) setKafka(viper *viper.Viper, port, secondPort, zookeeperPort, ip string) {

	d.setZookeeper(viper, zookeeperPort)

	serviceName := "kafka"
	servicePre := d.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", "wurstmeister/kafka")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".hostname", containerName)
	//viper.Set("services.kafkaserver.restart", "always")
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.Kafka, secondPort + ":" + d.InPort.KafkaSecond})

	viper.Set(servicePre+".environment.KAFKA_LISTENERS", fmt.Sprintf("INSIDE://:%v,OUTSIDE://:%v", d.InPort.Kafka, secondPort))
	viper.Set(servicePre+".environment.KAFKA_INTER_BROKER_LISTENER_NAME", "INSIDE")
	viper.Set(servicePre+".environment.KAFKA_ADVERTISED_LISTENERS",
		fmt.Sprintf("INSIDE://%v:%v,OUTSIDE://%v:%v", containerName, d.InPort.Kafka, ip, secondPort))
	viper.Set(servicePre+".environment.KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", "INSIDE:PLAINTEXT,OUTSIDE:PLAINTEXT")
	viper.Set(servicePre+".environment.KAFKA_ZOOKEEPER_CONNECT", d.getContainerName("zookeeper")+":"+d.InPort.Zookeeper)

}

func (d *Compose) setZookeeper(viper *viper.Viper, port string) {

	serviceName := "zookeeper"
	servicePre := d.getServicePre(serviceName)
	containerName := d.getContainerName(serviceName)

	viper.Set(servicePre+".image", "wurstmeister/zookeeper:latest")
	viper.Set(servicePre+".container_name", containerName)
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.Zookeeper})
}

func (d *Compose) setRedis(viper *viper.Viper, port string, registryCommon *string) {

	serviceName := "redis"
	servicePre := d.getServicePre(serviceName)

	viper.Set(servicePre+".image", d.GetImage("redis:3.2.11", registryCommon))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".hostname", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.Redis})
}

func (d *Compose) setNginx(viper *viper.Viper, project *Project, port string, registryCommon *string) {

	serviceName := "nginx"
	servicePre := d.getServicePre(serviceName)
	deps := []string{d.getServiceServer(project.Name)}

	viper.Set(servicePre+".image", d.GetImage("nginx:1.16", registryCommon))
	viper.Set(servicePre+".container_name", d.getContainerName(serviceName))
	viper.Set(servicePre+".ports", []string{port + ":" + d.InPort.Nginx})
	viper.Set(servicePre+".restart", "always")
	viper.Set(servicePre+".depends_on", deps)
	viper.Set(servicePre+".volumes", []string{
		"./nginx/default.conf:/etc/nginx/conf.d/default.conf",
		"./nginx/html:/usr/share/nginx/html",
	})

}

func (d *Compose) getServicePre(serviceName string) string {
	return "services." + d.getServiceServer(serviceName)
}

func (Compose) getServiceServer(serviceName string) string {
	return strings.ToLower(serviceName) + SUFSERVER
}

func (d *Compose) dependsOn(depends []string) []string {
	newDeps := make([]string, 0)
	for _, dep := range depends {
		newDeps = append(newDeps, d.getServiceServer(dep))
	}
	return newDeps
}

func (Compose) getContainerName(serviceName string) string {
	return PRETEST + serviceName
}
func (d *Compose) appCompose(viper *viper.Viper, project *Project) {
	servicePre := d.getServicePre(project.Name)

	viper.Set(servicePre+".image", project.Setting.Image+":latest")
	viper.Set(servicePre+".restart", "always")
	viper.Set(servicePre+".container_name", d.getContainerName(project.Name))
	viper.Set(servicePre+".environment", project.Setting.Envs)
	viper.Set(servicePre+".ports", project.Setting.Ports)

	viper.Set(servicePre+".depends_on", project.DependsOn)
}

func (d *Compose) setEventBroker(viper *viper.Viper, project *Project, imageEnv, ip string) {
	d.setEventProducer(viper, project.Owner.EventProducer, imageEnv, ip)
	d.setEventConsumer(viper, project.Owner.EventConsumer, project.Owner.StreamNames, imageEnv)
}
func (d *Compose) setEventProducer(viper *viper.Viper, project *Project, imageEnv, ip string) {
	servicePre := d.getServicePre(project.Name)
	viper.Set(servicePre+".image", "registry.p2shop.com.cn/"+project.Name+"-"+imageEnv+":latest")

	viper.Set(servicePre+".container_name", d.getContainerName(project.Name))
	viper.Set(servicePre+".environment", project.Setting.Envs)
	viper.Set(servicePre+".ports", project.Setting.Ports)

	viper.Set(servicePre+".depends_on", []string{d.getServiceServer("kafka")})
}
func (d *Compose) setEventConsumer(viper *viper.Viper, project *Project, streamNames []string, imageEnv string) {
	for _, sName := range streamNames {
		envs := append(project.Setting.Envs, "CONSUMER_GROUP_ID="+sName)

		name := project.Name + "-" + sName
		servicePre := d.getServicePre(name)
		viper.Set(servicePre+".image", "registry.p2shop.com.cn/"+project.Name+"-"+sName+"-"+imageEnv+":latest")

		viper.Set(servicePre+".container_name", d.getContainerName(name))
		viper.Set(servicePre+".environment", envs)
		viper.Set(servicePre+".depends_on", []string{
			d.getServiceServer("kafka"),
			d.getServiceServer("mysql"),
			d.getServiceServer("redis"),
		})
	}
}

func (d *Compose) upperKafkaEnvEland(ymlStr string) string {
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

func (d *Compose) upperKafkaEnv(ymlStr string) string {
	ymlStr = strings.Replace(ymlStr, "kafka_advertised_listeners", "KAFKA_ADVERTISED_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_inter_broker_listener_name", "KAFKA_INTER_BROKER_LISTENER_NAME", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listener_security_protocol_map", "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_listeners", "KAFKA_LISTENERS", -1)
	ymlStr = strings.Replace(ymlStr, "kafka_zookeeper_connect", "KAFKA_ZOOKEEPER_CONNECT", -1)

	ymlStr = strings.Replace(ymlStr, "kafka_advertised_port", "KAFKA_ADVERTISED_PORT", -1)
	return ymlStr
}

func (d *Compose) checkAll(project *Project, flag *Flag, dockercompose string) (err error) {
	ip := *flag.HostIp

	if project.Owner.IsMysql {
		if err = d.checkMysql(dockercompose, *flag.MysqlPort, ip); err != nil {
			return
		}
	}
	if project.Owner.IsSqlServer {
		if err = d.checkSqlServer(dockercompose, *flag.SqlServerPort, ip); err != nil {
			return
		}
	}
	if project.Owner.IsKafka {
		if err = d.checkKafka(dockercompose, *flag.KafkaPort, ip); err != nil {
			return
		}
	}
	return
}

func (d *Compose) checkMysql(dockercompose, port, ip string) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", MYSQL.String()+SUFSERVER); err != nil {
		return err
	}
	if err := d.DailMysql(port, ip); err != nil {
		return err
	}
	return nil
}
func (d *Compose) DailMysql(port, ip string) (err error) {
	dbType := MYSQL.String()
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
		break
	}
	if err != nil {
		return err
	}
	Info("finish ping " + dbType)
	return nil
}
func (d *Compose) checkSqlServer(dockercompose, port, ip string) (err error) {

	dbType := SQLSERVER.String()

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", dbType+SUFSERVER); err != nil {
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

func (d *Compose) checkKafka(dockercompose, port, ip string) (err error) {
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "zookeeper"+SUFSERVER); err != nil {
		return
	}

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "kafka"+SUFSERVER); err != nil {
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

func (d *Compose) dailKafka(port, ip string) (err error) {
	_, err = kafkautil.DialLeader(context.Background(), "tcp", ip+":"+port, "ping", 0)
	if err != nil {
		return
	}
	return
}
