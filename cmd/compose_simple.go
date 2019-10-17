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
	if BoolPointCheck(flag.DockerNoLogin) == false {
		r, err := Project{}.GetRegistryCommon(*flag.JwtToken)
		if err != nil {
			return err
		}
		Info("==> docker login " + r.Registry + " ...")
		if _, err := CmdRealtime("docker", "login", "-u", r.LoginName, "-p", r.Pwd, r.Registry); err != nil {
			return err
		}
	}
	dockercompose := fmt.Sprintf("%v/docker-compose.yml", TEMP_FILE)
	if BoolPointCheck(flag.DockerNoDown) == false {
		if err := d.Down(dockercompose, flag); err != nil {
			return err
		}
		Info("==> compose downed!")
	}
	if BoolPointCheck(flag.DockerNoCheck) == false {
		if err := d.CheckAll(serviceName, dockercompose, flag); err != nil {
			return err
		}
	}
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

	ip := *flag.DockerHostIp
	compose := &Compose{}
	compose.SetPort()
	serviceList := strings.Split(serviceName, ",")
	if ContainString(serviceList, KAFKASERVER.String()) {
		if err := CheckHost(ip, *flag.Prefix); err != nil {
			return err
		}
	}
	for _, name := range serviceList {
		switch name {
		case KAFKASERVER.String():
			compose.setKafkaEland(viper, *flag.DockerKafkaPort, ip, flag.RegistryCommon, *flag.Prefix)
		case MYSQLSERVER.String():
			mysqlPorts := strings.Split(*flag.DockerMysqlPort, ",")
			for i, mysqPort := range mysqlPorts {
				prefix := fmt.Sprint(*flag.Prefix, i, "-")
				compose.setMysqlSimple(viper, mysqPort, flag.RegistryCommon, prefix)
			}
		case SQLSERVERSERVER.String():
			compose.setSqlServer(viper, *flag.DockerSqlServerPort, flag.RegistryCommon, *flag.Prefix)
		case REDISSERVER.String():
			compose.setRedis(viper, *flag.DockerRedisPort, flag.RegistryCommon, *flag.Prefix)
		}
	}
	return compose.WriteYml(viper)
}

func (d ComposeSimple) CheckAll(serviceName, dockercompose string, flag *Flag) error {
	compose := Compose{}
	serviceList := strings.Split(serviceName, ",")
	ip := *flag.DockerHostIp
	for _, name := range serviceList {
		switch name {
		case KAFKASERVER.String():
			if err := compose.checkKafka(dockercompose, *flag.DockerKafkaPort, ip); err != nil {
				return err
			}
		case MYSQLSERVER.String():
			mysqlPorts := strings.Split(*flag.DockerMysqlPort, ",")
			for i, mysqPort := range mysqlPorts {
				prefix := fmt.Sprint(*flag.Prefix, i, "-")
				if err := compose.checkMysqlSimple(dockercompose, mysqPort, ip, prefix); err != nil {
					return err
				}
			}

		case SQLSERVERSERVER.String():
			if err := compose.checkSqlServer(dockercompose, *flag.DockerSqlServerPort, ip); err != nil {
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
