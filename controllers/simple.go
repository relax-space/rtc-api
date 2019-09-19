package controllers

import (
	"net/http"
	"rtc-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
)

type DbAccountApiController struct {
}

func (d DbAccountApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("", d.GetDbAccount)
}

func (d DbAccountApiController) GetDbAccount(c echo.Context) error {
	account, err := models.DbAccount{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, account)
}

type ImageAccountApiController struct {
}

func (d ImageAccountApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("", d.GetAll)
}

func (d ImageAccountApiController) GetAll(c echo.Context) error {
	accounts, err := models.ImageAccount{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, accounts)
}

type NamespaceApiController struct {
}

func (d NamespaceApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("", d.GetAll)
}

func (d NamespaceApiController) GetAll(c echo.Context) error {
	namespace, err := models.Namespace{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, namespace)
}

type TenantApiController struct {
}

func (d TenantApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("", d.GetAll)
}

func (d TenantApiController) GetAll(c echo.Context) error {
	tenant, err := models.Tenant{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, tenant)
}
