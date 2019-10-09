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

func (d ComposeSimple) Start(serviceName string, flag *Flag) error {
	if err := os.MkdirAll(TEMP_FILE, os.ModePerm); err != nil {
		return err
	}
	if err := d.ComposeYml(serviceName, flag); err != nil {
		return err
	}
	if BoolPointCheck(flag.NoLogin) == false {
		r, err := Project{}.GetRegistryCommon()
		if err != nil {
			return err
		}
		Info("==> docker login " + r.Registry + " ...")
		if _, err := CmdRealtime("docker", "login", "-u", r.LoginName, "-p", r.Pwd, r.Registry); err != nil {
			return err
		}
	}
	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	if err := d.Down(dockercompose, flag); err != nil {
		return err
	}
	if err := d.CheckAll(serviceName, dockercompose, flag); err != nil {
		return err
	}
	Info("==> compose downed!")

	if err := d.Up(dockercompose); err != nil {
		return err
	}
	Info(`==> compose up!`)
	return nil
}

func (d ComposeSimple) ComposeYml(serviceName string, flag *Flag) error {
	viper := viper.New()
	viper.Set("version", "3")
	viper.SetConfigName(YMLNAMEDOCKERCOMPOSE)
	viper.AddConfigPath(TEMP_FILE)

	ip := *flag.HostIp
	compose := &Compose{}
	compose.SetPort()
	serviceList := strings.Split(serviceName, ",")
	if ContainString(serviceList, KAFKASERVER.String()) {
		if err := CheckHost(ip); err != nil {
			return err
		}
	}
	for _, name := range serviceList {
		switch name {
		case KAFKASERVER.String():
			compose.setKafkaEland(viper, *flag.KafkaPort, ip, flag.RegistryCommon)
		case MYSQLSERVER.String():
			compose.setMysql(viper, *flag.MysqlPort, flag.RegistryCommon)
		case SQLSERVERSERVER.String():
			compose.setSqlServer(viper, *flag.SqlServerPort, flag.RegistryCommon)
		case REDISSERVER.String():
			compose.setRedis(viper, *flag.RedisPort, flag.RegistryCommon)
		}
	}
	return compose.WriteYml(viper)
}

func (d ComposeSimple) CheckAll(serviceName, dockercompose string, flag *Flag) error {
	compose := Compose{}
	serviceList := strings.Split(serviceName, ",")
	ip := *flag.HostIp
	for _, name := range serviceList {
		switch name {
		case KAFKASERVER.String():
			if err := compose.checkKafka(dockercompose, *flag.KafkaPort, ip); err != nil {
				return err
			}
		case MYSQLSERVER.String():
			if err := compose.checkMysql(dockercompose, *flag.MysqlPort, ip); err != nil {
				return err
			}
		case SQLSERVERSERVER.String():
			if err := compose.checkSqlServer(dockercompose, *flag.SqlServerPort, ip); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d ComposeSimple) Down(dockercompose string, flag *Flag) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "down", "--remove-orphans", "-v"); err != nil {
		return err
	}
	return nil
}

func (d ComposeSimple) Up(dockercompose string) error {
	if _, err := CmdRealtime("docker-compose", "-f", dockercompose, "up", "-d", "--no-recreate"); err != nil {
		return err
	}
	return nil
}
