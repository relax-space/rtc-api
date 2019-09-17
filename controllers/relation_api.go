package controllers

import (
	"github.com/pangpanglabs/echoswagger"
)

type RelationApiController struct {
}

// localhost:8080/docs
func (d RelationApiController) Init(g echoswagger.ApiGroup) {
	g.SetSecurity("Authorization")
	// g.GET("", d.GetAll).
	// 	AddParamQueryNested(SearchInput{})
	// g.GET("/:id", d.GetOne).
	// 	AddParamPath("", "id", "id").AddParamQuery("", "with_store", "with_store", false)
	// g.PUT("/:id", d.Update).
	// 	AddParamPath("", "id", "id").
	// 	AddParamBody(models.Relation{}, "relation", "only can modify name,color,price", true)
	// g.POST("", d.Create).
	// 	AddParamBody(models.Relation{}, "relation", "new relation", true)
	// g.DELETE("/:id", d.Delete).
	// 	AddParamPath("", "id", "id")
}
