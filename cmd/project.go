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
	Name      string `json:"name"` //service + "-" + namespace
	Service   string `json:"service"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`

	SubIds      []int       `json:"subIds"` //subIds
	Setting     *SettingDto `json:"setting"`
	Registry    string      `json:"registry"`
	RegistryPwd string      `json:"registryPwd"`

	Children  []*Project      `json:"children"`
	DependsOn []string        `json:"-"`
	Owner     ProjectOwnerDto `json:"-"`
}

type ProjectOwnerDto struct {
	IsKafka     bool
	IsMysql     bool
	IsSqlServer bool
	IsRedis     bool
	IsStream    bool

	DbTypes          []string
	ChildNames       []string
	StreamNames      []string
	EventProducer    *Project
	EventConsumer    *Project
	MysqlAccount     DbAccount
	SqlServerAccount DbAccount
}

type DbAccount struct {
	Host    string
	Port    int
	User    string
	Pwd     string
	DbNames []string
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
func (d Project) GetDbAccount(dbType DateBaseType) (DbAccount, error) {
	urlStr := fmt.Sprintf("%v/v1/dbaccounts/%v", env.RtcApiUrl, dbType.String())
	var resp struct {
		Success   bool      `json:"success"`
		DbAccount DbAccount `json:"result"`
	}
	statusCode, err := httpreq.New(http.MethodGet, urlStr, nil).WithToken(d.token()).Call(&resp)
	if err != nil {
		return DbAccount{}, err
	}
	if statusCode != http.StatusOK {
		return DbAccount{}, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
	}
	return resp.DbAccount, nil
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
	if statusCode != http.StatusOK {
		return int64(0), nil, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
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

func (d Project) GetEventProducer() (*Project, error) {
	return d.GetProject("kafka-producer")
}

func (d Project) GetEventConsumer() (*Project, error) {
	return d.GetProject("kafka-consumer")
}
