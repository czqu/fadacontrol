package utils

import "os"

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func FileCanRead(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsPermission(err)
}
func DirCanWrite(dirPath string) bool {
	_, err := os.Stat(dirPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return false
		}
		return true
	}
	return false

}
