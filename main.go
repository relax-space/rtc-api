package main

import (
	"log"

	"github.com/ElandGroup/joblog"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func main() {
	serviceName, flag := (Flag{}).Init()
	if StringPointCheck(serviceName) == false {
		return
	}
	if BoolPointCheck(flag.Debug) {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetFlags(0)
	}
	if BoolPointCheck(flag.NoLog) == false {
		log.Println("log init ...")
		initJobLog(serviceName, flag)
	}
	if comboResource = (ComboResource{}).GetInstance(flag.ComboResource, flag.RegistryCommon); comboResource == nil {
		Info("The --combo-resource parameter supports msl, srx, msl-srx. For details, see ./rtc run -h")
		return
	}
	if ContainString(EMPTYSERVER.List(), *serviceName) {
		port := Config{}.LoadFlagPort(flag)
		ComposeSimple{}.Start(*serviceName, "127.0.0.1", port, flag)
		return
	}

	c, err := Config{}.LoadEnv(*serviceName, flag)
	if err != nil {
		Error(err)
		return
	}
	if err = composeWriteYml(c); err != nil {
		return
	}
	if err = (Nginx{}).WriteConfig(c.Project); err != nil {
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

func initJobLog(serviceName *string, flag *Flag) {
	ip, err := currentIp()
	if err != nil {
		log.Println(jobLog.Err)
		return
	}

	jobLog = joblog.New(jobLogUrl, "rtc", map[string]interface{}{"service name:": serviceName, "ip": ip})
	if jobLog.Err != nil {
		log.Println(jobLog.Err)
		return
	}
	jobLog.Info(flag)
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
	e := EventBroker{}
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
	if e.ShouldEventBroker(c.Project) {
		streamNames := e.StreamList(c.Project)
		if err = (EventBroker{}).SetEventBroker(viper, c.Port.EventBroker, streamNames); err != nil {
			Error(err)
		}
	}

	d.setComposeApp(viper, c.Project)
	d.setComposeNginx(viper, c.Project, c.Port.Nginx)
	d.setComposeWaitStart(viper, c.Project)
	d.WriteYml(viper)
	return
}
