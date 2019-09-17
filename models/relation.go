package models

type Relation struct {
	Service         string               `json:"service"`
	Container       string               `json:"container"`
	Image           string               `json:"image"`
	GitlabShortName string               `json:"gitlabShortName"`
	Entrypoint      string               `json:"entrypoint"`
	Children        map[string]*Relation `json:"children"`

	Namespace      string              `json:"namespace"`
	Envs           []string            `json:"envs"`
	IsProjectKafka bool                `json:"isProjectKafka"`
	Ports          []string            `json:"ports"`
	Databases      map[string][]string `json:"databases"`
	StreamNames    []string            `json:"streamNames"`
	ExecPath       string              `json:"execPath"`
}
