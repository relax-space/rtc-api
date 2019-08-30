package main

import (
	"fmt"

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
		compose.checkKafka(dockercompose, port.Kafka)
	case MYSQLSERVER.String():
		compose.setComposeMysql(viper, port.Mysql)
		compose.WriteYml(viper)
		d.Down(dockercompose, flag)
		compose.checkMysql(dockercompose, port.Mysql)
	case SQLSERVERSERVER.String():
		compose.setComposeSqlserver(viper, port.SqlServer)
		compose.WriteYml(viper)
		d.Down(dockercompose, flag)
		compose.checkSqlServer(dockercompose, port.SqlServer)
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

	return
}

func (d ComposeSimple) Up(dockercompose string, flag *Flag) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return err
	}
	return nil
}
