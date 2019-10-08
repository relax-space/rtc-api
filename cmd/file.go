package cmd

import (
	"io/ioutil"
	"os"
	"path"
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

func (d Folder) DeleteAll(folderPath string) error {

	if has := d.IsExist(folderPath); has == false {
		return nil
	}
	dir, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return err
	}
	for _, d := range dir {
		if err = os.RemoveAll(path.Join(folderPath, d.Name())); err != nil {
			return err
		}
	}
	return nil
}
