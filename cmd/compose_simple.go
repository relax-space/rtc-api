package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ComposeSimple struct {
}

func (d ComposeSimple) ShouldSimple(serviceName string) bool {
	if strings.Contains(serviceName, ",") {
		return true
	}
	if ContainString(EMPTYSERVER.List(), serviceName) {
		return true
	}
	return false
}

func (d ComposeSimple) Start(serviceName, ip string, port PortDto, flag *Flag) error {
	if err := (File{}).DeleteAll("./" + TEMP_FILE + "/"); err != nil {
		return err
	}
	// create directory
	if err := os.MkdirAll(TEMP_FILE, os.ModePerm); err != nil {
		return err
	}
	if err := d.ComposeYml(serviceName, ip, port); err != nil {
		return err
	}
	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	if err := d.Down(dockercompose, flag); err != nil {
		return err
	}
	if err := d.CheckAll(serviceName, ip, dockercompose, port); err != nil {
		return err
	}
	if err := d.Up(dockercompose); err != nil {
		return err
	}
	Info(`==> compose up!`)
	return nil
}

func (d ComposeSimple) ComposeYml(serviceName, ip string, port PortDto) error {
	viper := viper.New()
	viper.Set("version", "3")
	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)

	compose := Compose{}
	serviceList := strings.Split(serviceName, ",")
	if ContainString(serviceList, KAFKASERVER.String()) {
		if err := CheckHost(ip); err != nil {
			return err
		}
	}
	for _, name := range serviceList {
		switch name {
		case KAFKASERVER.String():
			compose.setComposeKafkaEland(viper, port.Kafka, port.KafkaSecond, port.Zookeeper, ip)
		case MYSQLSERVER.String():
			compose.setComposeMysql(viper, port.Mysql)
		case SQLSERVERSERVER.String():
			compose.setComposeSqlserver(viper, port.SqlServer)
		case REDISSERVER.String():
			compose.setComposeRedis(viper, port.Redis)
		}
	}
	return compose.WriteYml(viper)
}

func (d ComposeSimple) CheckAll(serviceName, ip, dockercompose string, port PortDto) error {
	compose := Compose{}
	serviceList := strings.Split(serviceName, ",")
	for _, name := range serviceList {
		switch name {
		case KAFKASERVER.String():
			if err := compose.checkKafka(dockercompose, port.Kafka, ip); err != nil {
				return err
			}
		case MYSQLSERVER.String():
			if err := compose.checkMysql(dockercompose, port.Mysql, ip); err != nil {
				return err
			}
		case SQLSERVERSERVER.String():
			if err := compose.checkSqlServer(dockercompose, port.SqlServer, ip); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d ComposeSimple) Down(dockercompose string, flag *Flag) error {
	if BoolPointCheck(flag.NoLogin) == false {
		Info("==> docker login " + comboResource.Registry + " ...")
		if _, err := CmdRealtime("docker", "login", "-u", "eland", "-p", registryPwd, comboResource.Registry); err != nil {
			return err
		}
	}
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
		return err
	}
	Info("==> compose downed!")
	return nil
}

func (d ComposeSimple) Up(dockercompose string) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return err
	}
	return nil
}
