package utilfile

import "os"

func EnsureDirExits(dir string) error {
	ok, err := Exist(dir)
	if !ok && err == nil {
		err = os.MkdirAll(dir, os.ModePerm)
	}
	return err
}
