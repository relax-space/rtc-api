package controllers

const (
	DefaultMaxResultCount = 30
)

type SearchInput struct {
	Sortby         []string `query:"sortby"`
	Order          []string `query:"order"`
	SkipCount      int      `query:"skipCount"`
	MaxResultCount int      `query:"maxResultCount"`
	Like           string   `query:"like"`
}

type ProjectDto struct {
	Id         int    `json:"id"`
	Name       string `json:"name"` //service + "-" + tenantName + "-" + namespace
	Service    string `json:"service"`
	Namespace  string `json:"namespace"`
	TenantName string `json:"tenantName"`

	SubIds  []int      `json:"subIds"` //subIds
	Setting SettingDto `json:"setting"`
}

type SettingDto struct {
	Image          string              `json:"image"`
	Envs           []string            `json:"envs"`
	IsProjectKafka bool                `json:"isProjectKafka"`
	Ports          []string            `json:"ports"`
	Databases      map[string][]string `json:"databases"`
	StreamNames    []string            `json:"streamNames"`
}
