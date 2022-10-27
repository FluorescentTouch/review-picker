package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const path = "users.txt"

type File struct {
	f *os.File
}

func New() (*File, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &File{f: f}, nil
}

func (f *File) Close() error {
	return f.f.Close()
}

func (f *File) Read(dst any) error {
	b, err := ioutil.ReadAll(f.f)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dst)
}

func (f *File) Write(src any) error {
	if src == nil {
		return nil
	}
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	err = f.f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.f.WriteAt(b, 0)
	return err
}
