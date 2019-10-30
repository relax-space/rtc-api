package config

import (
	configutil "github.com/pangpanglabs/goutils/config"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/sirupsen/logrus"
)

var config C

func Init(appEnv string, options ...func(*C)) C {
	if err := configutil.Read(appEnv, &config); err != nil {
		logrus.WithError(err).Warn("Fail to load config file")
	}
	config.Appenv = appEnv

	for _, option := range options {
		option(&config)
	}

	return config
}

func Config() C {
	return config
}

type C struct {
	Database struct {
		Driver     string
		Connection string
		Logger     struct {
			Kafka echomiddleware.KafkaConfig
		}
	}
	RedisConn   string
	BehaviorLog struct {
		Kafka echomiddleware.KafkaConfig
	}
	EventBroker struct {
		Kafka echomiddleware.KafkaConfig
	}
	Services struct {
		OfferApi   string
		CslBrand   string
		CslProduct string
		RslProduct string
	}
	TccHooks map[string][]struct {
		Try, Confirm, Cancel string
	}
	Appenv, JwtSecret     string
	ServiceName, HttpPort string
	Debug                 bool
}
