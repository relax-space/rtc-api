package controllers_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"rtc-api/config"
	"rtc-api/models"
	"nomni/utils/auth"
	"nomni/utils/validator"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pangpanglabs/goutils/behaviorlog"
	configutil "github.com/pangpanglabs/goutils/config"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/pangpanglabs/goutils/jwtutil"
)

var (
	echoApp          *echo.Echo
	handleWithFilter func(handlerFunc echo.HandlerFunc, c echo.Context) error
)

func TestMain(m *testing.M) {
	db := enterTest()
	code := m.Run()
	exitTest(db)
	os.Exit(code)
}

func enterTest() *xorm.Engine {
	configutil.SetConfigPath("../")
	c := config.Init(os.Getenv("APP_ENV"), func(c *config.C) {
		if s := os.Getenv("JWT_SECRET"); s != "" {
			c.JwtSecret = s
			jwtutil.SetJwtSecret(s)
		}
	})
	xormEngine, err := xorm.NewEngine(c.Database.Driver, c.Database.Connection)
	if err != nil {
		panic(err)
	}
	if err = models.DropTables(xormEngine); err != nil {
		panic(err)
	}
	if err = models.InitTable(xormEngine); err != nil {
		panic(err)
	}

	echoApp = echo.New()
	echoApp.Validator = validator.New()
	db := echomiddleware.ContextDB("test", xormEngine, echomiddleware.KafkaConfig{})
	jwt := middleware.JWT([]byte(os.Getenv("JWT_SECRET")))
	behaviorlogger := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			c.SetRequest(req.WithContext(context.WithValue(req.Context(),
				behaviorlog.LogContextName, behaviorlog.New("test", req),
			)))
			return next(c)
		}
	}
	handleWithFilter = func(handlerFunc echo.HandlerFunc, c echo.Context) error {
		return behaviorlogger(jwt(auth.UserClaimMiddleware()(db(handlerFunc))))(c)
	}
	return xormEngine
}

func exitTest(db *xorm.Engine) {
	// if err := models.DropTables(db); err != nil {
	// 	panic(err)
	// }
}

func setHeader(r *http.Request) {
	token, _ := jwtutil.NewToken(map[string]interface{}{"aud": "colleague", "tenantCode": "test"})

	r.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
}
