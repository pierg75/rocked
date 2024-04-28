package utils

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func CleanupChrootDir(path string, create bool) (err error) {
	_, error := os.Stat(path)
	if error == os.ErrNotExist {
		return nil
	} else if error != nil {
		return error
	}
	os.RemoveAll(path)
	if create {
		os.Mkdir(path, 0644)
	}
	return nil
}

func ExtractImage(archive, dest string) error {
	reader, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer reader.Close()
	tar := tar.NewReader(reader)
	for {
		header, err := tar.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		path := filepath.Join(dest, header.Name)
		entry_info := header.FileInfo()
		if entry_info.IsDir() {
			err := os.MkdirAll(path, entry_info.Mode())
			if err != nil {
				return err
			}
			continue
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, entry_info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tar)
		if err != nil {
			return err
		}

	}
	return nil
}

func IsRoot() bool {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("Error trying to get the current user (%v)", err)
	}
	return user.Username == "root"
}
