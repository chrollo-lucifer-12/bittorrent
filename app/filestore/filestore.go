package filestore

import (
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
