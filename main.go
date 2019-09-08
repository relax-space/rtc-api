package main

import (
	"log"

	"github.com/ElandGroup/joblog"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func main() {
	isContinue, serviceName, flag := (Flag{}).Init()
	if isContinue==false {
		return
	}
	if StringPointCheck(serviceName) == false {
		panic("service name is required.")
	}
	if BoolPointCheck(flag.Debug) {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetFlags(0)
	}
	ip := getIp(flag.HostIp)
	if BoolPointCheck(flag.NoLog) == false {
		log.Println("log init ...")
		if err := initJobLog(serviceName, flag, ip); err != nil {
			log.Println(err)
			panic(err)
		}
	}
	if comboResource = (ComboResource{}).GetInstance(flag.ComboResource, flag.RegistryCommon); comboResource == nil {
		panic("The --combo-resource parameter supports p2shop, srx, p2shop-srx. For details, see ./rtc run -h")
	}
	// simple service
	if (ComposeSimple{}).ShouldSimple(*serviceName) {
		port := Config{}.LoadFlagPort(flag)
		if err := (ComposeSimple{}).Start(*serviceName, ip, port, flag); err != nil {
			Error(err)
			return
		}
		return
	}

	c, err := Config{}.LoadEnv(*serviceName, ip, flag)
	if err != nil {
		Error(err)
		return
	}
	if err = composeWriteYml(c, ip); err != nil {
		Error(err)
		return
	}
	if err = (Nginx{}).WriteConfig(c.Project); err != nil {
		Error(err)
		return
	}

	if err = writeLocal(c); err != nil {
		Error(err)
		return
	}

	if err = (Compose{}).Exec(c, flag, ip); err != nil {
		Error(err)
		return
	}

	Info("==> you can start testing now. check health by `docker ps -a`")
	return

}

func initJobLog(serviceName *string, flag *Flag, ip string) error {

	jobLog = joblog.New(jobLogUrl, "rtc", map[string]interface{}{"service name:": serviceName, "ip": ip})
	if jobLog.Err != nil {
		return jobLog.Err
	}
	jobLog.Info(flag)
	return nil
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

func composeWriteYml(c *FullDto, ip string) (err error) {
	viper := viper.New()
	p := ProjectInfo{}
	database := Database{}
	d := Compose{}
	e := EventBroker{}
	if p.ShouldKafka(c.Project) {
		if err = CheckHost(ip); err != nil {
			return
		}
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
			return
		}
	}

	d.setComposeApp(viper, c.Project)
	d.setComposeNginx(viper, c.Project, c.Port.Nginx)
	d.setComposeWaitStart(viper, c.Project)
	d.WriteYml(viper)
	return
}

func getIp(ipFlag *string) string {
	if StringPointCheck(ipFlag) {
		return *ipFlag
	}
	ip, err := currentIp()
	if err != nil {
		log.Println("WARNING: fetch ip failure, has set ip to 127.0.0.1")
		return "127.0.0.1"
	}
	return ip
}
