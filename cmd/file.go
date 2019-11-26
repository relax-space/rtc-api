package cmd

import (
	"io"
	"net/http"
	"github.com/pangpanglabs/goutils/httpreq"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type File struct {
}

func (File) WriteString(folderPath, fileName, content string) error {
	var err error
	if err = (Folder{}).MkdirAll(folderPath); err != nil {
		return err
	}
	fileName = path.Join(folderPath, fileName)
	var out *os.File
	if out, err = os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm); err != nil {
		return err
	}
	defer out.Close()
	if _, err = out.WriteString(content); err != nil {
		return err
	}
	return nil
}

func (File) WriteUrl(url, fileName, jwtToken string) (err error) {
	req := httpreq.New(http.MethodGet, url, nil, func(httpReq *httpreq.HttpReq) error {
		httpReq.RespDataType = httpreq.ByteArrayType
		return nil
	})
	resp, err := req.WithToken(jwtToken).RawCall()
	if err != nil {
		return
	}
	defer resp.Body.Close()

	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return
}

func (d File) Create(folderPath, fileName string) error {
	if err := (Folder{}).MkdirAll(folderPath); err != nil {
		return err
	}
	fileName = path.Join(folderPath, fileName)
	if _, err := os.Create(fileName); err != nil {
		return err
	}
	return nil
}

type Folder struct {
}

func (d Folder) MkdirAll(path string) error {
	if len(path) == 0 {
		return nil
	}
	if d.IsExist(path) {
		return nil
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (Folder) IsExist(fileName string) bool {
	s, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func (d Folder) Delete(folderPath, excludePath string) error {

	if has := d.IsExist(folderPath); has == false {
		return nil
	}
	dir, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return err
	}
	for _, v := range dir {
		name := path.Join(folderPath, v.Name())
		if d.IsShouldDelete(name, excludePath) {
			if err = os.RemoveAll(name); err != nil {
				return err
			}
		}

	}
	return nil
}

func (d Folder) DeleteAndIgnoreLocalSql(folderPath string, dbNet *string) error {
	exclude := ""
	if StringPointCheck(dbNet) && *dbNet == LOCALDBNET.String() {
		exclude = path.Join(TEMP_FILE, "database")
	}
	if err := d.Delete(folderPath, exclude); err != nil {
		return err
	}
	return nil
}

func (d Folder) IsShouldDelete(folderPath, excludePath string) bool {
	if len(excludePath) == 0 {
		return true
	}
	if strings.HasPrefix(folderPath, excludePath) {
		return false
	}
	return true
}
