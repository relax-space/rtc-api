package models_test

import (
	"fmt"
	"rtc-api/models"
	"testing"

	"github.com/pangpanglabs/goutils/test"
)

func TestCmdBiz(t *testing.T) {
	expId := 1
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

		SubIds: nil,
		Setting: models.SettingDto{
			Image:          "registry.p2shop.com.cn/go-api",
			Envs:           []string{"APP_ENV=rtc"},
			IsProjectKafka: false,
			Ports:          []string{"8080"},
			Databases:      map[string][]string{"mysql": []string{"fruit"}},
			StreamNames:    nil,
		},
	}
	expProject.Name = models.Project{}.GetName(expProject.TenantName, expProject.Namespace, expProject.Service)
	subProject1.Name = models.Project{}.GetName(subProject1.TenantName, subProject1.Namespace, subProject1.Service)

	projects := []models.Project{
		expProject,
		subProject1,
	}
	for i, project := range projects {
		t.Run(fmt.Sprint("Create#", i+1), func(t *testing.T) {
			f := &project
			affectedRow, err := f.Create(ctx)
			test.Ok(t, err)
			test.Equals(t, int64(1), affectedRow)
			test.Assert(t, f.Id == i+1, "create failure")
		})
	}

	t.Run("GetByName", func(t *testing.T) {
		has, v, err := models.Project{}.GetByName(ctx, expProject.Name)
		test.Ok(t, err)
		test.Equals(t, true, has)
		test.Equals(t, expId, v.Id)
		test.Equals(t, expProject.Name, v.Name)
		test.Equals(t, expProject.Service, v.Service)
		test.Equals(t, expProject.Namespace, v.Namespace)

		test.Equals(t, expProject.SubIds, v.SubIds)
		test.Equals(t, expProject.Setting, v.Setting)
	})
	t.Run("GetByIds", func(t *testing.T) {
		projects, err := models.Project{}.GetByIds(ctx, []int{1})
		test.Ok(t, err)
		v := projects[0]
		test.Equals(t, expId, v.Id)
		test.Equals(t, expProject.Name, v.Name)
		test.Equals(t, expProject.Service, v.Service)
		test.Equals(t, expProject.Namespace, v.Namespace)

		test.Equals(t, expProject.SubIds, v.SubIds)
		test.Equals(t, expProject.Setting, v.Setting)
	})

	t.Run("GetAll", func(t *testing.T) {
		totalCount, projects, err := models.Project{}.GetAll(ctx, nil, nil, 0, 1, "")
		test.Ok(t, err)
		v := projects[0]
		test.Equals(t, int64(2), totalCount)
		test.Equals(t, expId, v.Id)
		test.Equals(t, expProject.Name, v.Name)
		test.Equals(t, expProject.Service, v.Service)
		test.Equals(t, expProject.Namespace, v.Namespace)

		test.Equals(t, expProject.SubIds, v.SubIds)
		test.Equals(t, expProject.Setting, v.Setting)
	})
	t.Run("Update", func(t *testing.T) {
		service := "go-api5"
		f := &expProject
		f.Id = 1
		f.Service = service
		affectedRow, err := f.Update(ctx, f.Id)
		test.Ok(t, err)
		test.Equals(t, int64(1), affectedRow)
		test.Assert(t, f.Service == service, "update failure")
	})
	t.Run("Delete#2", func(t *testing.T) {
		id := 2
		affectedRow, err := models.Project{}.Delete(ctx, id)
		test.Ok(t, err)
		test.Equals(t, int64(1), affectedRow)
	})
}
