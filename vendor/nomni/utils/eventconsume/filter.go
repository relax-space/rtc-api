package eventconsume

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"nomni/utils/auth"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/ctxdb"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/kafka"

	"github.com/go-xorm/xorm"
	"github.com/labstack/gommon/random"
	"github.com/sirupsen/logrus"
)

type FilterFunc func(ctx context.Context) error
type Filter func(next FilterFunc) FilterFunc

func Handle(ctx context.Context, filters []Filter, f FilterFunc) error {
	for i := range filters {
		f = filters[len(filters)-1-i](f)
	}
	return f(ctx)
}

func ContextDB(service string, xormEngine *xorm.Engine, kafkaConfig kafka.Config) Filter {
	return ContextDBWithName(service, echomiddleware.ContextDBName, xormEngine, kafkaConfig)
}

func ContextDBWithName(service string, contexDBName echomiddleware.ContextDBType, xormEngine *xorm.Engine, kafkaConfig kafka.Config) Filter {
	ctxdb := ctxdb.New(xormEngine, service, kafkaConfig)

	return func(next FilterFunc) FilterFunc {
		return func(ctx context.Context) error {
			session := ctxdb.NewSession(ctx)
			defer session.Close()

			ctx = context.WithValue(ctx, contexDBName, session)

			return next(ctx)
		}
	}
}

func Recover() Filter {
	stackSize := 4 << 10 // 4 KB
	disableStackAll := false
	disablePrintStack := false

	return func(next FilterFunc) FilterFunc {
		return func(ctx context.Context) error {

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, stackSize)
					length := runtime.Stack(stack, !disableStackAll)
					if !disablePrintStack {
						log.Printf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					}
					behaviorlog.FromCtx(ctx).WithError(err)
				}
			}()

			return next(ctx)
		}
	}
}

func BehaviorLogger(serviceName string, config kafka.Config) Filter {
	var producer *kafka.Producer
	if config.Brokers != nil && config.Topic != "" {
		if p, err := kafka.NewProducer(config.Brokers, config.Topic,
			kafka.WithDefault(),
			kafka.WithTLS(config.SSL)); err != nil {
			logrus.Error("Create Kafka Producer Error", err)
		} else {
			producer = p
		}
	}

	return func(next FilterFunc) FilterFunc {
		return func(ctx context.Context) (err error) {
			behaviorLogContext := behaviorlog.FromCtx(ctx)
			behaviorLogContext.Producer = producer
			behaviorLogContext.Service = serviceName

			if behaviorLogContext.RequestID == "" {
				behaviorLogContext.RequestID = random.String(32)
			}

			if behaviorLogContext.ActionID == "" {
				behaviorLogContext.ActionID = random.String(32)
			}

			behaviorLogContext.Timestamp = time.Now()
			behaviorLogContext.RemoteIP = "127.0.0.1"
			behaviorLogContext.Host = "127.0.0.1"

			err = next(context.WithValue(ctx, behaviorlog.LogContextName, behaviorLogContext))
			if err != nil {
				behaviorLogContext.Err = err.Error()
			}

			behaviorLogContext.Write()

			return err
		}
	}
}

func UserClaimMiddleware() Filter {
	return func(next FilterFunc) FilterFunc {
		userClaim := auth.UserClaim{}
		return func(ctx context.Context) error {

			token := behaviorlog.FromCtx(ctx).AuthToken

			si := strings.Index(token, ".")
			li := strings.LastIndex(token, ".")
			if si == -1 || li == -1 || si == li {
				logrus.WithField("token", token).Warn("Invalid token")
				return next(ctx)
			}

			payload := token[si+1 : li]
			if payload == "" {
				logrus.WithField("token", token).Warn("Invalid token")
				return next(ctx)
			}

			payloadBytes, err := decodeSegment(payload)
			if err != nil {
				logrus.WithField("token", token).Warn("Invalid token")
				return next(ctx)
			}

			if err := json.Unmarshal(payloadBytes, &userClaim); err != nil {
				logrus.WithField("token", token).Warn("Invalid token")
				return next(ctx)
			}

			if userClaim.TenantCode == "" {
				logrus.WithField("token", token).Warn("Invalid token")
				return next(ctx)
			}
			switch userClaim.Iss {
			case auth.IssColleague:
				userClaim.ColleagueId = userClaim.Id
			case auth.IssMembership:
				userClaim.CustomerId = userClaim.Id
			}

			return next(context.WithValue(ctx, "userClaim", userClaim))
		}
	}
}

func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}
