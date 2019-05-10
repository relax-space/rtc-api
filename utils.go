package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/pangpanglabs/goutils/httpreq"

	"github.com/spf13/viper"
)

func Cmd(name string, arg ...string) (result string, err error) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("err:%v--stderr:%v", err, stderr.String())
		return
	}
	result = out.String()
	if len(result) != 0 {
		fmt.Println(result)
	}
	return
}

func CmdRealtime(name string, arg ...string) (result string, err error) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return
	}

	return
}

// func CmdRealtime(name string, args ...string) (result string, err error) {
// 	cmd := exec.Command(name, args...)
// 	var stdBuffer bytes.Buffer
// 	mw := io.MultiWriter(os.Stdout, &stdBuffer)

// 	cmd.Stdout = mw
// 	cmd.Stderr = mw

// 	// Execute the command
// 	if err = cmd.Run(); err != nil {
// 		return
// 	}
// 	result = stdBuffer.String()
// 	if len(result) != 0 {
// 		fmt.Println(result)
// 	}
// 	return
// }

// func CmdRealtime(name string, args ...string) (result string, err error) {
// 	cmd := exec.Command(name, args...)
// 	var stdBuffer bytes.Buffer
// 	mw := io.MultiWriter(os.Stdout, &stdBuffer)

// 	cmd.Stdout = mw
// 	cmd.Stderr = mw

// 	// Execute the command
// 	if err = cmd.Run(); err != nil {
// 		return
// 	}
// 	result = stdBuffer.String()
// 	if len(result) != 0 {
// 		fmt.Println(result)
// 	}
// 	return
// }

func createIfNot(path string) error {
	if _, err := os.Stat(path); err != nil {
		if _, err = os.Create(path); err != nil {
			return err
		}
	}
	return nil
}

func Read(env string, config interface{}) error {
	viper.SetConfigName("config")
	viper.AddConfigPath(TEMP_FILE)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}

	if env != "" {
		f, err := os.Open("config." + env + ".yml")
		if err != nil {
			return fmt.Errorf("Fatal error config file: %s \n", err)
		}
		defer f.Close()
		viper.MergeConfig(f)
	}

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("Fatal error config file: %s \n", err)
	}
	return nil
}

func fetchFromgitlab(url, privateToken string) (b []byte, err error) {
	req := httpreq.New(http.MethodGet, url, nil, func(httpReq *httpreq.HttpReq) error {
		httpReq.RespDataType = httpreq.ByteArrayType
		return nil
	})
	req.Req.Header.Set("PRIVATE-TOKEN", privateToken)
	resp, err := req.RawCall()
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("request gitlab error:%v", url)
		return
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	return
}

func fetchTofile(url, fileName, privateToken string) (err error) {
	req := httpreq.New(http.MethodGet, url, nil, func(httpReq *httpreq.HttpReq) error {
		httpReq.RespDataType = httpreq.ByteArrayType
		return nil
	})
	req.Req.Header.Set("PRIVATE-TOKEN", privateToken)
	resp, err := req.RawCall()
	if err != nil {
		return
	}
	defer resp.Body.Close()

	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return
}

func fetchSqlTofile(projectDto *ProjectDto, privateToken string) (err error) {
	urlString := fmt.Sprintf("%v/test_info%v/table.sql", projectDto.GitRaw, getPathWhenMulti(projectDto))
	filePath := fmt.Sprintf("%v/%v.sql", TEMP_FILE, projectDto.ServiceName)
	if err = fetchTofile(urlString, filePath, PRIVATETOKEN); err != nil {
		err = fmt.Errorf("download sql error,url:%v,err:%v", urlString, err)
		return
	}
	return
}

func writeFile(fileName, content string) (err error) {
	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = out.WriteString(content)
	return
}

type DateBaseType int

const (
	MYSQL DateBaseType = iota
	REDIS
	MONGO
	SQLSERVER
)

func (DateBaseType) List() []string {
	return []string{"mysql", "redis", "mongo", "sqlserver"}
}

func (d DateBaseType) String() string {
	return d.List()[d]
}

type ScopeType int

const (
	ALL ScopeType = iota
	LocalData
	NONE
)

func (ScopeType) List() []string {
	return []string{"all", "localdata", "none"}
}

func (d ScopeType) String() string {
	return d.List()[d]
}

//fileName="/foo/123_*"
func deleteFileRegex(fileName string) (err error) {
	files, err := filepath.Glob(fileName)
	if err != nil {
		return
	}
	for _, f := range files {
		if err = os.Remove(f); err != nil {
			return
		}
	}
	return
}

func getGoPath() (gopath string) {
	gopath = os.Getenv("GOPATH")
	gopath = strings.Replace(gopath, "\\", "/", -1)
	return
}

func deleteSlice(chars []string, name string) (newChars []string) {
	for i := len(chars) - 1; i >= 0; i-- {
		if chars[i] == name {
			chars = append(chars[:i], chars[i+1:]...)
		}
	}
	newChars = chars
	return
}

func ContainString(chars []string, name string) bool {
	for _, c := range chars {
		if strings.ToUpper(c) == strings.ToUpper(name) {
			return true
		}
	}
	return false
}

func yamlStringSettings(vip *viper.Viper) (ymlString string, err error) {
	c := vip.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		return
	}
	ymlString = string(bs)
	return
}

func inIps() (ips []string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	ips = make([]string, 0)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return
}

func getIp(ipParam *string) (currentIp string) {
	if ipParam != nil && len(*ipParam) != 0 {
		currentIp = *ipParam
		return
	}
	ips := inIps()
	for _, ip := range ips {
		if strings.HasPrefix(ip, "10.202.101.") {
			currentIp = ip
			break
		}
	}
	return
}

func getProjectEnv(projectDto *ProjectDto) (err error) {
	gitRaw := fmt.Sprintf("%v/%v/raw/%v", PREGITHTTPURL, projectDto.GitShortPath, app_env)
	urlString := fmt.Sprintf("%v/test_info%v/project.yml", gitRaw, getPathWhenMulti(projectDto))
	projectDto.GitRaw = gitRaw
	b, err := fetchFromgitlab(urlString, PRIVATETOKEN)
	if err != nil {
		err = fmt.Errorf("read project.yml error:%v,url:%v", err, urlString)
		return
	}
	if err = yaml.Unmarshal(b, projectDto); err != nil {
		err = fmt.Errorf("parse project.yml error,project:%v,err:%v", projectDto.ServiceName, err.Error())
		return
	}
	if projectDto.IsMulti && len(projectDto.Envs) == 0 {
		getProjectEnv(projectDto)
	}
	return
}

func getPathWhenMulti(projectDto *ProjectDto) (path string) {
	if projectDto.IsMulti {
		path += "/" + projectDto.ServiceName
	}
	return
}
