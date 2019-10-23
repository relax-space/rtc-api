package controllers

import (
	"net/http"
	"nomni/utils/api"
	"rtc-api/models"
	"strconv"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/echoswagger"
)

type ProjectApiController struct {
}

func (d ProjectApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	g.GET("", d.GetAll).
		AddParamQueryNested(SearchInput{}).
		AddParamQuery("", "name", "go-api", true).
		AddParamQuery("", "with_child", "true", false).
		AddParamQuery("", "simple", "true", false)
	g.GET("/:id", d.GetById).
		AddParamPath("", "id", "1").
		AddParamQuery("", "with_child", "true", false)
	g.POST("", d.Create).
		AddParamBody(models.Project{}, "project", "new project", true)
	g.PUT("/:id", d.Update).
		AddParamPath("", "id", "1").
		AddParamBody(models.Project{}, "project", "update project", true)
}

func (d ProjectApiController) GetAll(c echo.Context) error {
	if len(c.QueryParam("name")) != 0 {
		return d.GetByName(c)
	}
	if len(c.QueryParam("simple")) != 0 {
		return d.GetAllSimple(c)
	}
	var v SearchInput
	if err := c.Bind(&v); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.ParameterParsingError(err))
	}
	if v.MaxResultCount == 0 {
		v.MaxResultCount = DefaultMaxResultCount
	}
	totalCount, items, err := models.Project{}.GetAll(c.Request().Context(), v.Sortby, v.Order, v.SkipCount, v.MaxResultCount)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if len(items) == 0 {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotFoundError())
	}
	return ReturnApiSucc(c, http.StatusOK, api.ArrayResult{
		TotalCount: totalCount,
		Items:      items,
	})

}
func (d ProjectApiController) GetById(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.InvalidParamError("id", c.Param("id"), err))
	}
	has, project, err := models.Project{}.GetById(c.Request().Context(), int(id))
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if has == false {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotFoundError())
	}
	if status, err := d.getWithChild(c, project); err != nil {
		return ReturnApiFail(c, status, err)
	}

	return ReturnApiSucc(c, http.StatusOK, project)
}
func (d ProjectApiController) GetAllSimple(c echo.Context) error {
	project, err := models.Project{}.GetAllSimple(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	return ReturnApiSucc(c, http.StatusOK, project)
}
func (d ProjectApiController) GetByName(c echo.Context) error {
	name := c.QueryParam("name")
	has, project, err := models.Project{}.GetByName(c.Request().Context(), name)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if has == false {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotFoundError())
	}

	if status, err := d.getWithChild(c, project); err != nil {
		return ReturnApiFail(c, status, err)
	}
	return ReturnApiSucc(c, http.StatusOK, project)
}
func (ProjectApiController) Filter(ids []int, projects []*models.Project) []*models.Project {
	pFilters := make([]*models.Project, 0)
	for _, id := range ids {
		for _, p := range projects {
			if id == p.Id {
				pFilters = append(pFilters, p)
			}
		}
	}
	return pFilters
}

func (d ProjectApiController) loopGet(c echo.Context, project *models.Project, projects []*models.Project) {
	if len(project.SubIds) != 0 {
		subProjects := d.Filter(project.SubIds, projects)
		project.Children = subProjects
		for k, v := range project.Children {
			if len(v.SubIds) != 0 {
				d.loopGet(c, project.Children[k], projects)
			}
		}
	}
}

func (d ProjectApiController) Create(c echo.Context) error {
	var v models.Project
	if err := c.Bind(&v); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.ParameterParsingError(err))
	}
	if err := c.Validate(v); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.ParameterParsingError(err))
	}
	v.Name = d.GetName(v.TenantName, v.Namespace, v.Service)
	has, _, err := models.Project{}.GetByName(c.Request().Context(), v.Name)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if has {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotCreatedError())
	}
	affectedRow, err := v.Create(c.Request().Context())
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if affectedRow == int64(0) {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotCreatedError())
	}
	return ReturnApiSucc(c, http.StatusCreated, v)
}

func (d ProjectApiController) Update(c echo.Context) error {

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.InvalidParamError("id", c.Param("id"), err))
	}

	var v models.Project
	if err := c.Bind(&v); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.ParameterParsingError(err))
	}
	if err := c.Validate(v); err != nil {
		return ReturnApiFail(c, http.StatusBadRequest, api.ParameterParsingError(err))
	}
	v.Id = int(id)
	v.Name = d.GetName(v.TenantName, v.Namespace, v.Service)
	has, _, err := models.Project{}.GetById(c.Request().Context(), v.Id)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if has == false {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotUpdatedError())
	}
	affectedRow, err := v.Update(c.Request().Context(), v.Id)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if affectedRow == int64(0) {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotUpdatedError())
	}
	return ReturnApiSucc(c, http.StatusOK, v)
}

func (d ProjectApiController) getWithChild(c echo.Context, project *models.Project) (int, error) {
	var withChild bool
	var err error
	if len(c.QueryParam("with_child")) != 0 {
		withChild, err = strconv.ParseBool(c.QueryParam("with_child"))
		if err != nil {
			return http.StatusBadRequest, api.InvalidParamError("with_child", c.Param("with_child"), err)
		}
	}
	if withChild == true {
		items, err := models.Project{}.GetAllReal(c.Request().Context())
		if err != nil {
			return http.StatusInternalServerError, err
		}
		d.loopGet(c, project, items)
	}
	if err := (ProjectOwner{}).Reload(c.Request().Context(), project); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (d ProjectApiController) GetName(tenantName, namespace, service string) string {
	namespaceNew := models.Project{}.SetName(tenantName, namespace)
	name := models.Project{}.SetName(service, namespaceNew)
	return name

}
