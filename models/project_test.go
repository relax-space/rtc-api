package models_test

import (
	"rtc-api/models"
	"testing"

	"github.com/pangpanglabs/goutils/test"
)

func TestCmdBiz(t *testing.T) {
	expProject := models.Project{
		Service:    "go-api",
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
	expProject.Name = models.Project{}.SetName(expProject.Service, expProject.Namespace)
	t.Run("Create", func(t *testing.T) {
		f := &expProject
		affectedRow, err := f.Create(ctx)
		test.Ok(t, err)
		test.Equals(t, int64(1), affectedRow)
		test.Assert(t, f.Id == int(1), "create failure")
	})
	t.Run("GetByName", func(t *testing.T) {
		has, v, err := models.Project{}.GetByName(ctx, expProject.Name)
		test.Ok(t, err)
		test.Equals(t, true, has)
		test.Equals(t, expProject.Id, v.Id)
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
		test.Equals(t, expProject.Id, v.Id)
		test.Equals(t, expProject.Name, v.Name)
		test.Equals(t, expProject.Service, v.Service)
		test.Equals(t, expProject.Namespace, v.Namespace)

		test.Equals(t, expProject.SubIds, v.SubIds)
		test.Equals(t, expProject.Setting, v.Setting)
	})

	t.Run("GetAll", func(t *testing.T) {
		totalCount, projects, err := models.Project{}.GetAll(ctx, nil, nil, 0, 1)
		test.Ok(t, err)
		v := projects[0]
		test.Equals(t, int64(1), totalCount)
		test.Equals(t, expProject.Id, v.Id)
		test.Equals(t, expProject.Name, v.Name)
		test.Equals(t, expProject.Service, v.Service)
		test.Equals(t, expProject.Namespace, v.Namespace)

		test.Equals(t, expProject.SubIds, v.SubIds)
		test.Equals(t, expProject.Setting, v.Setting)
	})

}
