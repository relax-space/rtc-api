package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type ComposeSimple struct {
}

func (d ComposeSimple) Start(serviceName, ip string, port PortDto, flag *Flag) error {
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
		if err := d.Down(dockercompose, flag); err != nil {
			return err
		}
		if err := compose.checkKafka(dockercompose, port.Kafka, ip); err != nil {
			return err
		}
	case MYSQLSERVER.String():
		compose.setComposeMysql(viper, port.Mysql)
		compose.WriteYml(viper)
		if err := d.Down(dockercompose, flag); err != nil {
			return err
		}
		if err := compose.checkMysql(dockercompose, port.Mysql, ip); err != nil {
			return err
		}
	case SQLSERVERSERVER.String():
		compose.setComposeSqlserver(viper, port.SqlServer)
		compose.WriteYml(viper)
		if err := d.Down(dockercompose, flag); err != nil {
			return err
		}
		if err := compose.checkSqlServer(dockercompose, port.SqlServer, ip); err != nil {
			return err
		}
	case REDISSERVER.String():
		compose.setComposeSqlserver(viper, port.Redis)
		compose.WriteYml(viper)
		if err := d.Down(dockercompose, flag); err != nil {
			return err
		}
		if err := d.Up(dockercompose, flag); err != nil {
			return err
		}
	}

	Info(`==> compose up!`)
	return nil
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
