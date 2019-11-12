package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/pangpanglabs/goutils/httpreq"
)

type Project struct {
	Id         int    `json:"id"`
	Name       string `json:"name"` //service + "-" + namespace
	Service    string `json:"service"`
	Namespace  string `json:"namespace"`
	TenantName string `json:"tenantName"`

	SubIds  []int       `json:"subIds"` //subIds
	Setting *SettingDto `json:"setting"`

	Children  []*Project      `json:"children"`
	DependsOn []string        `json:"dependsOn"`
	Owner     ProjectOwnerDto `json:"owner"`
}

type ProjectOwnerDto struct {
	IsKafka     bool
	IsMysql     bool
	IsSqlServer bool
	IsRedis     bool
	IsStream    bool

	DbTypes       []string
	ChildNames    []string
	StreamNames   []string
	EventProducer *Project
	EventConsumer *Project
	Databases     map[string][]DatabaseDto
	ImageAccounts []ImageAccountDto
}

type DatabaseDto struct {
	TenantName string
	Namespace  string
	DbName     string
}

type DbAccountDto struct {
	TenantName string
	Name       string
	Host       string
	Port       int
	User       string
	Pwd        string
	THost      string
	TPort      int
	TUser      string
	TPwd       string
}
type ImageAccountDto struct {
	Registry  string `json:"registry"`
	LoginName string `json:"loginName"`
	Pwd       string `json:"pwd"`
}

type NamespaceDto struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type SettingDto struct {
	Image          string              `json:"image"`
	Envs           []string            `json:"envs"`
	IsProjectKafka bool                `json:"isProjectKafka"`
	Ports          []string            `json:"ports"`
	Databases      map[string][]string `json:"databases"`
	StreamNames    []string            `json:"streamNames"`
}

type PLoop struct {
	Children []*Project `json:"projects"`
}

func (d Project) GetServiceNames(q, jwtToken string) ([]string, error) {

	projects, err := d.GetAll(jwtToken)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	for _, p := range projects {
		list = append(list, p.Name)
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
	return newList, nil
}

func (d Project) GetProject(serviceName, jwtToken string) (*Project, error) {
	urlStr := fmt.Sprintf("%v/v1/projects?name=%v&depth=-1", env.RtcApiUrl, serviceName)
	var resp struct {
		Success bool     `json:"success"`
		Project *Project `json:"result"`
	}
	statusCode, err := httpreq.New(http.MethodGet, urlStr, nil).WithToken(jwtToken).Call(&resp)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
	}
	return resp.Project, nil
}
func (d Project) GetDbAccount(dbAccounts []DbAccountDto, dbType DateBaseType, tenant string) DbAccountDto {
	for _, dbAccount := range dbAccounts {
		if dbAccount.TenantName == tenant && dbAccount.Name == dbType.String() {
			return dbAccount
		}
	}
	return DbAccountDto{}
}
func (d Project) GetAllDbAccount(jwtToken string) ([]DbAccountDto, error) {
	urlStr := fmt.Sprintf("%v/v1/db_accounts", env.RtcApiUrl)
	var resp struct {
		Success   bool           `json:"success"`
		DbAccount []DbAccountDto `json:"result"`
	}
	statusCode, err := httpreq.New(http.MethodGet, urlStr, nil).WithToken(jwtToken).Call(&resp)
	if err != nil {
		return resp.DbAccount, err
	}
	if statusCode != http.StatusOK {
		return resp.DbAccount, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
	}
	return resp.DbAccount, nil
}

func (d Project) GetAll(jwtToken string) ([]*Project, error) {
	pLoop := &PLoop{
		Children: make([]*Project, 0),
	}
	if err := d.getLoop(0, 2000, jwtToken, pLoop); err != nil {
		return nil, err
	}
	return pLoop.Children, nil
}

func (d Project) GetImageAccount(jwtToken string) ([]ImageAccountDto, error) {
	urlStr := fmt.Sprintf("%v/v1/image_accounts", env.RtcApiUrl)
	var resp struct {
		Success       bool              `json:"success"`
		ImageAccounts []ImageAccountDto `json:"result"`
	}
	statusCode, err := httpreq.New(http.MethodGet, urlStr, nil).WithToken(jwtToken).Call(&resp)
	if err != nil {
		return resp.ImageAccounts, err
	}
	if statusCode != http.StatusOK {
		return resp.ImageAccounts, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
	}
	return resp.ImageAccounts, nil
}

func (d Project) GetRegistryCommon(jwtToken string) (ImageAccountDto, error) {
	imageAccounts, err := d.GetImageAccount(jwtToken)
	if err != nil {
		return ImageAccountDto{}, err
	}
	for _, imageAccount := range imageAccounts {
		if imageAccount.Registry == REGISTRYCOMMON {
			return imageAccount, nil
		}
	}
	return ImageAccountDto{}, errors.New("no found common registry")
}

func (d Project) getLoop(skipCount, maxResultCount int64, jwtToken string, pLoop *PLoop) error {
	totalCount, pList, err := d.get(skipCount, maxResultCount, jwtToken)
	if err != nil {
		return err
	}
	pLoop.Children = append(pLoop.Children, pList...)
	if d.isContinueSync(skipCount, maxResultCount, totalCount) {
		skipCount = skipCount + maxResultCount
		d.getLoop(skipCount, maxResultCount, jwtToken, pLoop)
	}
	return nil
}

func (d Project) isContinueSync(skipCount, maxResultCount, totalCount int64) bool {
	if int64(skipCount+maxResultCount) < totalCount {
		return true
	}
	return false
}

func (d Project) get(skipCount, maxResultCount int64, jwtToken string) (int64, []*Project, error) {
	urlStr := fmt.Sprintf("%v/v1/projects?skipCount=%v&maxResultCount=%v", env.RtcApiUrl, skipCount, maxResultCount)
	var resp struct {
		Success     bool `json:"success"`
		ArrayResult struct {
			Items      []*Project `json:"items"`
			TotalCount int64      `json:"totalCount"`
		} `json:"result"`
	}
	statusCode, err := httpreq.New(http.MethodGet, urlStr, nil).WithToken(jwtToken).Call(&resp)
	if err != nil {
		return int64(0), nil, err
	}
	if statusCode != http.StatusOK {
		return int64(0), nil, fmt.Errorf("http status exp:200,act:%v,url:%v", statusCode, urlStr)
	}
	return resp.ArrayResult.TotalCount, resp.ArrayResult.Items, nil
}
