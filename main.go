package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func main() {

	flag.Parse()
	if ok := flagCheck(); !ok {
		return
	}

	c, err := Config{}.LoadEnv(envDto.ServiceName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = writeLocal(c); err != nil {
		fmt.Println(err)
		return
	}

	if err = (Compose{}).Exec(c); err != nil {
		fmt.Println(err)
		return
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt)
	go func() {
		for s := range signals {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				os.Exit(0)
			}
		}
	}()
	time.Sleep(100 * time.Hour)
}

func flagCheck() (ok bool) {
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	if envDto.V {
		fmt.Println("version:1.0.0")
		flag.Usage()
		return
	}

	if envDto.H {
		fmt.Println(msg)
		flag.Usage()
		return
	}

	if len(envDto.ServiceName) == 0 {
		fmt.Println(`-s is required.`)
		flag.Usage()
		return
	}
	ok = true
	return
}

func writeLocal(c *FullDto) (err error) {
	if scope == LOCAL.String() {
		return
	}
	if err = composeWriteYml(c); err != nil {
		return
	}
	if err = (Nginx{}).WriteConfig(c.Project, c.Port.EventBroker); err != nil {
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
	d := Compose{}
	if p.ShouldKafka(c.Project) {
		d.setComposeKafkaEland(viper, c.Port.Kafka, c.Port.KafkaSecond, c.Port.Zookeeper, c.Ip)
	}
	if p.ShouldDb(c.Project, MYSQL) {
		d.setComposeMysql(viper, c.Port.Mysql)
	}
	if p.ShouldDb(c.Project, REDIS) {
		d.setComposeRedis(viper, c.Port.Redis)
	}

	if p.ShouldEventBroker(c.Project) {
		streamNames := p.StreamList(c.Project)
		EventBroker{}.SetEventBroker(viper, c.Port.EventBroker, streamNames)
	}
	d.setComposeApp(viper, c.Project)
	d.setComposeNginx(viper, c.Project.ServiceName, c.Port.Nginx)
	d.WriteYml(viper)
	return
}
