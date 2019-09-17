package controllers

import (
	"nomni/utils/api"

	"github.com/pangpanglabs/goutils/behaviorlog"

	"github.com/labstack/echo"
)

func ReturnApiFail(c echo.Context, status int, err error) error {
	behaviorlog.FromCtx(c.Request().Context()).WithError(err)
	if apiError, ok := err.(api.Error); ok {
		return c.JSON(status, api.Result{
			Error: apiError,
		})
	}
	return c.JSON(status, api.Result{
		Success: false,
		Error:   api.UnknownError(err),
	})
}

func ReturnApiSucc(c echo.Context, status int, result interface{}) error {
	behaviorlog.FromCtx(c.Request().Context()).WithBizAttrs(map[string]interface{}{"resp": result})
	return c.JSON(status, api.Result{
		Success: true,
		Result:  result,
	})
}
