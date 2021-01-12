package tools

import "os"

func DirExists(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsDir() {
			return true
		}
	}
	return false
}

func FileExists(name string) bool {
	if fi, err := os.Stat(name); os.IsNotExist(err) {
		return false
	} else {
		return !fi.IsDir()
	}
}
