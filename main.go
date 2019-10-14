package main

import (
	"fmt"
	"log"
	"net/http"
	"nomni/utils/validator"
	"os"
	"rtc-api/cmd"
	"rtc-api/config"
	"rtc-api/controllers"
	"rtc-api/models"
	"strings"

	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/echomiddleware"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pangpanglabs/echoswagger"
	"github.com/sirupsen/logrus"
)
var Version string
func main() {
	isApi := os.Getenv("IS_RTC_API")
	if isApi != "Y" {
		cmd.Start(Version)
		return
	}

	c := config.Init(os.Getenv("APP_ENV"))

	fmt.Println("Config===", c)
	db, err := models.InitDB(c.Database.Driver, c.Database.Connection)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := models.InitTable(db); err != nil {
		panic(err)
	}

	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	r := echoswagger.New(e, "docs", &echoswagger.Info{
		Title:       "Sample Relation API",
		Description: "This is docs for relation service",
		Version:     "1.0.0",
	})
	r.AddSecurityAPIKey("Authorization", "JWT token", echoswagger.SecurityInHeader)
	r.SetUI(echoswagger.UISetting{
		HideTop: true,
	})
	controllers.ProjectApiController{}.Init(r.Group("projects", "v1/projects"))
	controllers.DbAccountApiController{}.Init(r.Group("db_accounts", "v1/db_accounts"))
	controllers.ImageAccountApiController{}.Init(r.Group("image_accounts", "v1/image_accounts"))
	controllers.NamespaceApiController{}.Init(r.Group("namespaces", "v1/namespaces"))
	controllers.TenantApiController{}.Init(r.Group("tenants", "v1/tenants"))

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Use(middleware.RequestID())
	e.Use(echomiddleware.ContextLogger())
	e.Use(echomiddleware.ContextDB(c.ServiceName, db, c.Database.Logger.Kafka))
	e.Use(echomiddleware.BehaviorLogger(c.ServiceName, c.BehaviorLog.Kafka))
	if !strings.HasSuffix(c.Appenv, "production") {
		behaviorlog.SetLogLevel(logrus.InfoLevel)
	}

	e.Validator = validator.New()

	e.Debug = c.Debug

	if err := e.Start(":" + c.HttpPort); err != nil {
		log.Println(err)
	}

}
