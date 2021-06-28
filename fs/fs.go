package fs

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	defaultTimeout = 10 * time.Second
	discardLogger  = log.New(ioutil.Discard, "", 0)
)

func lookupDirs(extraDirs []string) []string {
	pathVar := os.Getenv("PATH")
	dirs := filepath.SplitList(pathVar)
	for _, ep := range extraDirs {
		dirs = append(dirs, ep)
	}
	return dirs
}

type fileCheckFunc func(path string) error

func findFile(dirs []string, file string, f fileCheckFunc) (string, error) {
	for _, dir := range dirs {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		path := filepath.Join(dir, file)
		if err := f(path); err == nil {
			return path, nil
		}
	}
	return "", exec.ErrNotFound
}

func checkExecutable(file string) error {
	d, err := os.Stat(file)
	if err != nil {
		return err
	}
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return nil
	}
	return os.ErrPermission
}
