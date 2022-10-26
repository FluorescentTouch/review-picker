package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const path = "users.txt"

type File struct {
	*os.File
}

func Open() (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &File{File: f}, nil
}

func (f *File) Close() error {
	return f.File.Close()
}

func (f *File) Read(dst any) error {
	b, err := ioutil.ReadAll(f.File)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func (f *File) Write(src any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	_, err = f.File.Write(b)
	return err
}
