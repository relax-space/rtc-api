package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	kafkautil "github.com/segmentio/kafka-go"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type ComposeSimple struct {
}

func (d ComposeSimple) Start(serviceName, ip string, port PortDto, flag *Flag) {
	viper := viper.New()
	viper.Set("version", "3")
	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)

	compose := Compose{}
	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	switch serviceName {
	case KAFKASERVER.String():
		compose.setComposeKafkaEland(viper, port.Kafka, port.KafkaSecond, port.Zookeeper, ip)
		compose.WriteYml(viper)
		d.Down(dockercompose, flag)
		d.checkKafka(dockercompose, port.Kafka)
	case MYSQLSERVER.String():
		compose.setComposeMysql(viper, port.Mysql)
		compose.WriteYml(viper)
		d.Down(dockercompose, flag)
		d.checkMysql(dockercompose, port.Mysql)
	case SQLSERVERSERVER.String():
		compose.setComposeSqlserver(viper, port.SqlServer)
		compose.WriteYml(viper)
		d.Down(dockercompose, flag)
		d.checkSqlServer(dockercompose, port.SqlServer)
	case REDISSERVER.String():
		compose.setComposeSqlserver(viper, port.Redis)
		compose.WriteYml(viper)
		d.Down(dockercompose, flag)
		d.Up(dockercompose, flag)
	}

	Info(`==> compose up!`)
}
func (d ComposeSimple) Down(dockercompose string, flag *Flag) (err error) {

	if BoolPointCheck(flag.NoLogin) == false {
		Info("==> docker login " + comboResource.Registry + " ...")
		if _, err = CmdRealtime("docker", "login", "-u", "eland", "-p", registryPwd, comboResource.Registry); err != nil {
			fmt.Printf("err:%v", err)
			return
		}
	}

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
		return
	}
	Info("==> compose downed!")
	return
}

func (d ComposeSimple) Up(dockercompose string, flag *Flag) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return err
	}
	return nil
}

func (d ComposeSimple) checkMysql(dockercompose, port string) (err error) {
	dbType := MYSQL.String()

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", dbType+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	Info("begin ping " + dbType + ",127.0.0.1:" + port)
	db, err := sql.Open("mysql", fmt.Sprintf("root:1234@tcp(127.0.0.1:%v)/mysql?charset=utf8", port))
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

func (d ComposeSimple) checkSqlServer(dockercompose, port string) (err error) {

	dbType := SQLSERVER.String()

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", dbType+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}
	Info("begin ping " + dbType + ",127.0.0.1:" + port)
	db, err := sql.Open("sqlserver",
		fmt.Sprintf("sqlserver://sa:Eland123@127.0.0.1:%v?database=master", port))

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

func (d ComposeSimple) checkKafka(dockercompose, port string) (err error) {
	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "zookeeper"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	if _, err = CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate", "kafka"+SUFSERVER); err != nil {
		fmt.Printf("err:%v", err)
		return
	}

	Info("begin ping kafka,127.0.0.1:" + port)
	for index := 1; index < 300; index++ {
		if err = d.dailKafka(port); err != nil {
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

func (d ComposeSimple) dailKafka(port string) (err error) {
	_, err = kafkautil.DialLeader(context.Background(), "tcp", "127.0.0.1:"+port, "ping", 0)
	if err != nil {
		return
	}
	return
}
