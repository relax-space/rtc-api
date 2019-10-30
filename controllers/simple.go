package controllers

import (
	"net/http"
	"nomni/utils/api"
	"rtc-api/models"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
)

type FrontApiController struct {
}

func (d FrontApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("", d.GetForEdit).
		AddParamQuery("", "tenantName", "pangpang", false)
}

func (d FrontApiController) GetForEdit(c echo.Context) error {
	tenants, err := models.Tenant{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	tenantName := c.QueryParam("tenantName")
	if len(tenantName) == 0 {
		return ReturnApiFail(c, http.StatusBadRequest, api.MissRequiredParamError("name"))
	}
	namespaces, err := models.Namespace{}.GetByTenantName(c.Request().Context(), tenantName)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}

	imageAccounts, err := models.ImageAccount{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	simpleProjects, err := models.Project{}.GetAllSimple(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	editMap := map[string]interface{}{
		"tenants":        tenants,
		"namespaces":     namespaces,
		"imageAccounts":  imageAccounts,
		"simpleProjects": simpleProjects,
	}
	return ReturnApiSucc(c, http.StatusOK, editMap)
}

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
	g.GET("/:name/namespaces", d.GetNsByTenantName).AddParamPath("", "name", "pangpang")
}

func (d TenantApiController) GetAll(c echo.Context) error {
	tenant, err := models.Tenant{}.GetAll(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, tenant)
}

func (d TenantApiController) GetNsByTenantName(c echo.Context) error {
	tenantName := c.Param("name")
	if len(tenantName) == 0 {
		return ReturnApiFail(c, http.StatusBadRequest, api.MissRequiredParamError("name"))
	}
	namespace, err := models.Namespace{}.GetByTenantName(c.Request().Context(), tenantName)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, namespace)
}
