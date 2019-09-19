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
		AddParamQueryNested(SearchInput{})
	g.GET("/:name", d.GetProject).
		AddParamPath("", "name", "go-api").
		AddParamQuery("", "with_child", "true", false)
	g.POST("", d.Create).
		AddParamBody(models.Project{}, "project", "new project", true)
}

func (ProjectApiController) GetAll(c echo.Context) error {
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

func (d ProjectApiController) GetProject(c echo.Context) error {
	name := c.Param("name")
	has, project, err := models.Project{}.GetByName(c.Request().Context(), name)
	if err != nil {
		return ReturnApiFail(c, http.StatusInternalServerError, err)
	}
	if has == false {
		return ReturnApiFail(c, http.StatusBadRequest, api.NotFoundError())
	}

	var withChild bool
	if len(c.QueryParam("with_child")) != 0 {
		withChild, err = strconv.ParseBool(c.QueryParam("with_child"))
		if err != nil {
			return ReturnApiFail(c, http.StatusBadRequest, api.InvalidParamError("with_child", c.Param("with_child"), err))
		}
	}
	if withChild == true {
		items, err := models.Project{}.GetAllReal(c.Request().Context())
		if err != nil {
			return ReturnApiFail(c, http.StatusInternalServerError, err)
		}
		d.loopGet(c, project, items)
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
	v.Name = models.Project{}.SetName(v.Service, v.Namespace)
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
