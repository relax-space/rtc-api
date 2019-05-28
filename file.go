package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/pangpanglabs/goutils/httpreq"
	"github.com/spf13/viper"
)

type File struct {
}

func (File) IsExist(fileName string) (isExist bool, err error) {
	if _, err = os.Stat(fileName); err != nil {
		if os.IsNotExist(err) == false {
			return
		}
		err = nil
		return
	}
	isExist = true
	return
}

func (File) WriteString(fileName, content string) (err error) {
	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	_, err = out.WriteString(content)
	return
}

func (d File) WriteViper(path string, viper *viper.Viper) (err error) {
	// create directory
	if err = os.MkdirAll(TEMP_FILE, os.ModePerm); err != nil {
		return
	}
	//create file
	if err = d.CreateEmpty(path); err != nil {
		return
	}
	if err = viper.WriteConfig(); err != nil {
		return
	}
	return
}

func (File) CreateEmpty(path string) (err error) {
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) == false {
			return
		}
		if _, err = os.Create(path); err != nil {
			return
		}
	}
	return
}

func (File) WriteUrl(url, fileName, privateToken string) (err error) {
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

func (File) ReadUrl(url, privateToken string) (b []byte, err error) {
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
		err = fmt.Errorf("status:%v", resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	return
}

func (File) DeleteRegex(fileName string) (err error) {
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

func (File) CreateSafe(path string) (err error) {
	isExist, err := File{}.IsExist(path)
	if err != nil {
		return
	}
	if isExist == false {
		if _, err = os.Create(path); err != nil {
			return
		}
	}
	return
}

func (File) DeleteAll(fileName string) (err error) {
	has, err := File{}.IsExist(fileName)
	if err != nil {
		return
	}
	if has == false {
		return
	}
	dir, err := ioutil.ReadDir(fileName)
	for _, d := range dir {
		if err = os.RemoveAll(path.Join([]string{fileName, d.Name()}...)); err != nil {
			return
		}
	}
	return
}

func (File) ReadViper(env string, config interface{}) error {
	viper.SetConfigName(YMLNAMECONFIG)
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
