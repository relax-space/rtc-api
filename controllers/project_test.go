package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"nomni/utils/api"
	"rtc-api/controllers"
	"rtc-api/models"
	"testing"

	"github.com/labstack/echo"
	"github.com/pangpanglabs/goutils/test"
)

func TestCmdBiz(t *testing.T) {
	expProject := models.Project{
		Service:    "go-api",
		Namespace:  "",
		TenantName: "",

		SubIds: []int{2},
		Setting: models.SettingDto{
			Image:          "registry.p2shop.com.cn/go-api",
			Envs:           []string{"APP_ENV=rtc"},
			IsProjectKafka: false,
			Ports:          []string{"8080"},
			Databases:      map[string][]string{"mysql": []string{"fruit"}},
			StreamNames:    nil,
		},
	}
	subProject1 := models.Project{
		Service:    "go-api2",
		Namespace:  "",
		TenantName: "",
		SubIds:     []int{3, 4},
		Setting: models.SettingDto{
			Image:     "registry.p2shop.com.cn/go-api",
			Envs:      []string{"APP_ENV=rtc"},
			Ports:     []string{"8080"},
			Databases: map[string][]string{"sqlserver": []string{"fruit"}},
		},
	}
	subProject2 := models.Project{
		Service:    "go-api3",
		Namespace:  "",
		TenantName: "",
		SubIds:     nil,
		Setting: models.SettingDto{
			Image:          "registry.p2shop.com.cn/go-api",
			IsProjectKafka: true,
			Envs:           []string{"APP_ENV=rtc"},
			Ports:          []string{"8080"},
			Databases:      map[string][]string{"redis": []string{"fruit"}},
		},
	}
	subProject3 := models.Project{
		Service:    "go-api4",
		Namespace:  "",
		TenantName: "",
		SubIds:     nil,
		Setting: models.SettingDto{
			Image:          "registry.p2shop.com.cn/go-api",
			IsProjectKafka: true,
			Envs:           []string{"APP_ENV=rtc"},
			Ports:          []string{"8080"},
			Databases:      map[string][]string{"mysql": []string{"fruit"}},
		},
	}
	expProject.Name = models.Project{}.GetName(expProject.TenantName, expProject.Namespace, expProject.Service)
	newProjects := []models.Project{
		expProject,
		subProject1,
		subProject2,
		subProject3,
	}
	for i, p := range newProjects {
		pb, _ := json.Marshal(p)
		t.Run(fmt.Sprint("Create#", i+1), func(t *testing.T) {
			expCreateProject := p
			expCreateProject.Id = i + 1
			expCreateProject.Name = models.Project{}.GetName(expCreateProject.TenantName, expCreateProject.Namespace, expCreateProject.Service)

			req := httptest.NewRequest(echo.POST, "/v1/projects", bytes.NewReader(pb))
			rec := httptest.NewRecorder()
			test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.Create, echoApp.NewContext(req, rec)))
			test.Equals(t, http.StatusCreated, rec.Code)
			var resp struct {
				Success bool           `json:"success"`
				Result  models.Project `json:"result"`
			}
			test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			test.Equals(t, true, resp.Success)
			test.Equals(t, expCreateProject, resp.Result)
		})
	}

	t.Run("GetAll", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/projects?sortby=id&order=asc&skipCount=0&maxResultCount=1", nil) //go-api is the second data,because orderby id desc
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.GetAll, c))
		test.Equals(t, http.StatusOK, rec.Code)
		var resp struct {
			Success bool `json:"success"`
			Result  struct {
				TotalCount int64            `json:"totalCount"`
				Items      []models.Project `json:"items"`
			} `json:"result"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, true, resp.Success)
		test.Equals(t, int64(4), resp.Result.TotalCount)
		test.Equals(t, 1, len(resp.Result.Items))
		result := resp.Result.Items[0]
		test.Equals(t, expProject.Name, result.Name)
		test.Equals(t, expProject.Service, result.Service)
		test.Equals(t, expProject.Namespace, result.Namespace)

		test.Equals(t, expProject.SubIds, result.SubIds)
		test.Equals(t, expProject.Setting, result.Setting)
		test.Equals(t, expProject.Children, result.Children)

	})

	t.Run("GetAllLike", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/v1/projects?like=api4", nil) //go-api is the second data,because orderby id desc
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.GetAll, c))
		test.Equals(t, http.StatusOK, rec.Code)
		var resp struct {
			Success bool `json:"success"`
			Result  struct {
				TotalCount int64            `json:"totalCount"`
				Items      []models.Project `json:"items"`
			} `json:"result"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, true, resp.Success)
		test.Equals(t, int64(1), resp.Result.TotalCount)
		test.Equals(t, 1, len(resp.Result.Items))
		result := resp.Result.Items[0]
		test.Equals(t, "go-api4", result.Name)

	})

	t.Run("GetById", func(t *testing.T) {
		id := 1
		req := httptest.NewRequest(echo.GET, "/?depth=-1", nil) //go-api is the second data,because orderby id desc
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/projects/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(id))
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.GetById, c))
		test.Equals(t, http.StatusOK, rec.Code)
		var resp struct {
			Result  models.Project `json:"result"`
			Success bool           `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, true, resp.Success)
		result := resp.Result
		test.Equals(t, expProject.Name, result.Name)
		test.Equals(t, expProject.Service, result.Service)
		test.Equals(t, expProject.Namespace, result.Namespace)

		test.Equals(t, expProject.SubIds, result.SubIds)
		test.Equals(t, expProject.Setting, result.Setting)
		test.Equals(t, 1, len(result.Children))
		test.Equals(t, "go-api2", result.Children[0].Service)
		test.Equals(t, "go-api3", result.Children[0].Children[0].Service)
		test.Equals(t, "go-api4", result.Children[0].Children[1].Service)

	})

	t.Run("GetByName", func(t *testing.T) {
		url := fmt.Sprintf("/v1/projects?name=%v&depth=-1", expProject.Name)
		req := httptest.NewRequest(echo.GET, url, nil)
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.GetAll, c))
		test.Equals(t, http.StatusOK, rec.Code)
		var resp struct {
			Result  models.Project `json:"result"`
			Success bool           `json:"success"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, true, resp.Success)
		result := resp.Result
		test.Equals(t, expProject.Name, result.Name)
		test.Equals(t, expProject.Service, result.Service)
		test.Equals(t, expProject.Namespace, result.Namespace)

		test.Equals(t, expProject.SubIds, result.SubIds)
		test.Equals(t, expProject.Setting, result.Setting)
		test.Equals(t, 1, len(result.Children))
		test.Equals(t, "go-api2", result.Children[0].Service)
		test.Equals(t, "go-api3", result.Children[0].Children[0].Service)
		test.Equals(t, "go-api4", result.Children[0].Children[1].Service)
	})
	t.Run("Update#1", func(t *testing.T) {
		id := 2
		expUpdateProject := subProject1
		expUpdateProject.Id = id
		expUpdateProject.Service = "go-api5"
		expUpdateProject.Name = models.Project{}.GetName(expUpdateProject.TenantName, expUpdateProject.Namespace, expUpdateProject.Service)
		pb, _ := json.Marshal(expUpdateProject)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewReader(pb))
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/projects/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(id))
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.Update, c))
		test.Equals(t, http.StatusOK, rec.Code)
		var resp struct {
			Success bool           `json:"success"`
			Result  models.Project `json:"result"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, true, resp.Success)
		test.Equals(t, expUpdateProject, resp.Result)

	})
	t.Run("Delete#1", func(t *testing.T) {
		id := 2
		req := httptest.NewRequest(echo.DELETE, "/", nil) //go-api is the second data,because orderby id desc
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/projects/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(id))
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.Delete, c))
		test.Equals(t, http.StatusBadRequest, rec.Code)
		var resp struct {
			Result  models.Project `json:"result"`
			Success bool           `json:"success"`
			Error   api.Error      `json:"error"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, false, resp.Success)
		test.Equals(t, "删除失败，有微服务 [1] 依赖于这个服务.", resp.Error.Error())
	})
	t.Run("Delete#2", func(t *testing.T) {
		id := 1
		req := httptest.NewRequest(echo.DELETE, "/", nil) //go-api is the second data,because orderby id desc
		rec := httptest.NewRecorder()
		c := echoApp.NewContext(req, rec)
		c.SetPath("/v1/projects/:id")
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprint(id))
		test.Ok(t, handleWithFilter(controllers.ProjectApiController{}.Delete, c))
		test.Equals(t, http.StatusOK, rec.Code)
		var resp struct {
			Result  models.Project `json:"result"`
			Success bool           `json:"success"`
			Error   api.Error      `json:"error"`
		}
		test.Ok(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		test.Equals(t, true, resp.Success)
		test.Equals(t, "", resp.Error.Error())
	})
}
