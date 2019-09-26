package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/pangpanglabs/goutils/httpreq"
)

type Project struct {
	Id        int    `json:"id"`
	Name      string `json:"name"` //service + "|" + namespace
	Service   string `json:"service"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`

	SubIds      []int       `json:"subIds"` //subIds
	Setting     *SettingDto `json:"setting"`
	Registry    string      `json:"registry"`
	RegistryPwd string      `json:"registryPwd"`

	Children  []*Project `json:"children"`
	DependsOn []string
	Owner     ProjectOwnerDto
}

type ProjectOwnerDto struct {
	IsKafka     bool
	IsMysql     bool
	IsSqlServer bool
	IsRedis     bool
	DbNames     []string
	Databases   map[string][]string
	ChildNames  []string
	StreamNames []string
}

type SettingDto struct {
	Entrypoint     string              `json:"entrypoint"`
	Envs           []string            `json:"envs"`
	IsProjectKafka bool                `json:"isProjectKafka"`
	Ports          []string            `json:"ports"`
	Databases      map[string][]string `json:"databases"`
	StreamNames    []string            `json:"streamNames"`
}

type PLoop struct {
	Children []*Project `json:"projects"`
}

func (d Project) GetServiceNames(q string)([]string,error){

	projects,err :=d.GetAll()
	if err != nil{
		return nil,err
	}
	list:= make([]string,0)
	for _, p := range projects {
		list= append(list,p.Name)
	}
	newList := make([]string, 0)
	if len(q) != 0 {
		for _, v := range list {
			vlow := strings.ToLower(v)
			if strings.Contains(vlow, strings.ToLower(q)) {
				newList = append(newList, v)
			}
		}
	} else {
		newList = list
	}
	return newList,nil
}

func (d Project) GetProject(serviceName string) (*Project, error) {
	urlStr := fmt.Sprintf("%v/v1/projects/%v?with_child=true", env.RtcApiUrl, serviceName)
	var resp struct {
		Success bool     `json:"success"`
		Project *Project `json:"result"`
	}
	statusCode,err:=httpreq.New(http.MethodGet, urlStr,nil).WithToken(d.token()).Call(&resp)
	if err != nil{
		return nil,err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
	}
	return resp.Project,nil
}
func (d Project) GetAll() ([]*Project, error) {
	pLoop := &PLoop{
		Children: make([]*Project, 0),
	}
	if err := d.getLoop(0, 2000, pLoop); err != nil {
		return nil, err
	}
	return pLoop.Children, nil
}

func (d Project) getLoop(skipCount,maxResultCount int64,pLoop *PLoop) error{
	totalCount,pList,err :=d.get(skipCount,maxResultCount)
	if err !=nil{
		return err
	}
	pLoop.Children = append(pLoop.Children, pList...)
	if d.isContinueSync(skipCount, maxResultCount, totalCount) {
		skipCount = skipCount + maxResultCount
		d.getLoop(skipCount,maxResultCount,pLoop)
	}
	return nil
}

func (d Project) isContinueSync(skipCount,maxResultCount,totalCount int64) bool {
		if int64(skipCount+maxResultCount) < totalCount {
			return true
		}
	return false
}

func (d Project) get(skipCount, maxResultCount int64) (int64, []*Project, error) {
	urlStr := fmt.Sprintf("%v/v1/projects?skipCount=%v&maxResultCount=%v", env.RtcApiUrl, skipCount, maxResultCount)
	var resp struct {
		Success     bool `json:"success"`
		ArrayResult struct {
			Items      []*Project `json:"items"`
			TotalCount int64      `json:"totalCount"`
		} `json:"result"`
	}
	statusCode,err:=httpreq.New(http.MethodGet, urlStr,nil).WithToken(d.token()).Call(&resp)
	if err != nil{
		return int64(0),nil,err
	}
	if statusCode != http.StatusOK{
		return int64(0),nil,fmt.Errorf("http status exp:200,act:%v",statusCode)
	}
	return resp.ArrayResult.TotalCount,resp.ArrayResult.Items,nil
}

func (d Project) token() string {
	token := os.Getenv("JWT_TOKEN")
	if len(token) == 0 {
		panic(errors.New("miss environment: JWT_TOKEN"))
	}
	return token
}

func (d *Project) LoadOwner() {
	d.ShouldKafka()
	d.Owner.IsMysql = d.ShouldDb(MYSQL)
	d.Owner.IsSqlServer = d.ShouldDb(SQLSERVER)
	d.Owner.IsRedis = d.ShouldDb(REDIS)
	d.Owner.Databases = d.Database()
	list := d.Database()
	d.Owner.DbNames = d.DatabaseList(list)
	d.SetNames()
	d.SetDependLoop()
	d.SetStreams()

}

func (d *Project) ShouldKafka() {
	d.kafka([]*Project{d})
	return
}

func (d *Project) ShouldDb(dbType DateBaseType) bool {
	list := d.Database()
	for k := range list {
		if dbType.String() == k {
			return true
		}
	}
	return false
}

func (d *Project) DatabaseList(list map[string][]string) []string {
	dbNames := make([]string, 0)
	for k := range list {
		dbNames = append(dbNames, k)
	}
	return dbNames
}

func (d *Project) Database() (list map[string][]string) {
	list = make(map[string][]string, 0)
	projects := []*Project{d}
	d.database(list, projects)
	for k, v := range list {
		list[k] = Unique(v)
	}
	return
}

func (d *Project) database(list map[string][]string, projects []*Project) {
	for _, project := range projects {
		for k, v := range project.Setting.Databases {
			if _, ok := list[k]; ok {
				list[k] = append(list[k], v...)
			} else {
				list[k] = v
			}
		}
		if len(project.Children) != 0 {
			d.database(list, project.Children)
		}
	}
	return
}

func (d *Project) kafka(projects []*Project) {
	d.Owner.IsKafka = false
	for _, project := range projects {
		if project.Setting.IsProjectKafka {
			d.Owner.IsKafka = true
			return
		}
		if len(project.Children) != 0 {
			d.kafka(project.Children)
		}
	}
	return
}

func (d *Project) SetNames() {
	d.childNames(d)
}

func (d *Project) childNames(pLoop *Project) {
	for _, v := range pLoop.Children {
		d.Owner.ChildNames = append(d.Owner.ChildNames, v.Name)
		if len(v.Children) != 0 {
			d.childNames(v)
		}
	}
}

func (d *Project) SetDependLoop() {
	Project{}.setDepend(d)
	Project{}.setDependLoop(d)
}

func (d Project) setDependLoop(pLoop *Project) {
	for k, v := range pLoop.Children {
		pLoop.DependsOn = append(pLoop.DependsOn, v.Name)
		Project{}.setDepend(pLoop.Children[k])
		if len(v.Children) != 0 {
			d.setDependLoop(pLoop.Children[k])
		}
	}
}
func (d Project) setDepend(pLoop *Project) {
	if pLoop.Setting.IsProjectKafka {
		pLoop.DependsOn = append(pLoop.DependsOn, "kafka")
	}
	dbNames := pLoop.DatabaseList(pLoop.Setting.Databases)
	pLoop.DependsOn = append(pLoop.DependsOn, dbNames...)
}

func (d *Project) SetStreams() {
	d.setStreams(d)
	d.Owner.StreamNames = append(d.Owner.StreamNames, d.Setting.StreamNames...)
	d.Owner.StreamNames = Unique(d.Owner.StreamNames)
}

func (d *Project) setStreams(pLoop *Project) {
	for _, v := range pLoop.Children {
		d.Owner.StreamNames = append(d.Owner.StreamNames, v.Setting.StreamNames...)
		if len(v.Children) != 0 {
			d.setStreams(v)
		}
	}
}
