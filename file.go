package lirity

import (
	"os"
	"path/filepath"
)

// DirSize 统计文件夹的大小
func DirSize(path string) (size int64, count int, err error) {
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return err
	})
	return
}
