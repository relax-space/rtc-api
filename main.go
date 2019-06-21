package main

import (
	"log"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func main() {
	serviceName, flag := (Flag{}).Init()
	if StringPointCheck(serviceName) == false {
		return
	}
	if BoolPointCheck(flag.Log) {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetFlags(0)
	}
	c, err := Config{}.LoadEnv(*serviceName, flag)
	if err != nil {
		Error(err)
		return
	}
	if err = composeWriteYml(c); err != nil {
		return
	}
	if err = (Nginx{}).WriteConfig(c.Project, c.Port.EventBroker); err != nil {
		return
	}

	if err = writeLocal(c); err != nil {
		Error(err)
		return
	}

	if err = (Compose{}).Exec(c, flag); err != nil {
		Error(err)
		return
	}

	Info("==> you can start testing now. check health by `docker ps -a`")
	return

}

func writeLocal(c *FullDto) (err error) {
	if scope == LOCAL.String() {
		return
	}
	if err = (ProjectInfo{}).WriteSql(c.Project); err != nil {
		return
	}

	if err = (Config{}).WriteYml(c); err != nil {
		return
	}
	return
}

func composeWriteYml(c *FullDto) (err error) {
	viper := viper.New()
	p := ProjectInfo{}
	database := Database{}
	d := Compose{}
	if p.ShouldKafka(c.Project) {
		d.setComposeKafkaEland(viper, c.Port.Kafka, c.Port.KafkaSecond, c.Port.Zookeeper, c.Ip)
	}
	if database.ShouldDbLoop(c.Project, MYSQL) {
		d.setComposeMysql(viper, c.Port.Mysql)
	}
	if database.ShouldDbLoop(c.Project, REDIS) {
		d.setComposeRedis(viper, c.Port.Redis)
	}
	if database.ShouldDbLoop(c.Project, SQLSERVER) {
		d.setComposeSqlserver(viper, c.Port.SqlServer)
	}

	if p.ShouldEventBroker(c.Project) {
		streamNames := p.StreamList(c.Project)
		if err = (EventBroker{}).SetEventBroker(viper, c.Port.EventBroker, streamNames); err != nil {
			Error(err)
		}
	}
	d.setComposeApp(viper, c.Project)
	d.setComposeNginx(viper, c.Project.ServiceName, c.Port.Nginx)
	d.WriteYml(viper)
	return
}
