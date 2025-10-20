package filestore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type FileStoreOpts struct {
	Dirname string
}

type FileStore struct {
	dirname string
}

func NewFileStore(opts FileStoreOpts) *FileStore {
	err := os.Mkdir(opts.Dirname, os.ModePerm)
	if err != nil {
		return nil
	}
	return &FileStore{
		dirname: opts.Dirname,
	}
}

func (f *FileStore) AddFile(filename string, data []byte) {
	filePath := filepath.Join(f.dirname, filename)
	outFile, _ := os.Create(filePath)
	outFile.Write(data)
	outFile.Close()
}

func (f *FileStore) CombineFiles() {
	files, err := os.ReadDir(f.dirname)
	if err != nil {
		fmt.Println(err)
		return
	}
	var b bytes.Buffer
	for _, file := range files {
		fileName := file.Name()
		filePath := filepath.Join(f.dirname, fileName)
		fileData, _ := os.ReadFile(filePath)
		b.Write(fileData)
		os.Remove(filePath)
	}

	filePath := filepath.Join(f.dirname, "sample.txt")
	outFile, _ := os.Create(filePath)
	outFile.Write(b.Bytes())
	outFile.Close()
}
