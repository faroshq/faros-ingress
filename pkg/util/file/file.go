package utilfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	dir, file := path.Split(filename)
	tempFile, err := ioutil.TempFile(dir, fmt.Sprintf(".%s", file))
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(data); err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempFile.Name(), perm); err != nil {
		return err
	}

	return os.Rename(tempFile.Name(), filename)
}

// Exist checks to see if a file exist at the provided path.
func Exist(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// ignore missing files
			return false, nil
		}
		return false, err
	}
	defer f.Close()
	return true, nil
}
