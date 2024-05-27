package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"log/slog"
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

// Unzip a file into a new unzipped file
func Gunzip(archive, dest string) error {
	gzippedFile, err := os.Open(archive)
	if err != nil {
		fmt.Printf("Error opening the file\n")
		return err
	}
	defer gzippedFile.Close()

	// Create a new gzip reader
	gzipReader, err := gzip.NewReader(gzippedFile)
	if err != nil {
		fmt.Printf("Error getting a new reader\n")
		return err
	}
	defer gzipReader.Close()

	// Create a new file to hold the uncompressed data
	uncompressedFile, err := os.Create(dest)
	if err != nil {
		fmt.Printf("Error creating dest\n")
		return err
	}
	defer uncompressedFile.Close()

	// Copy the contents of the gzip reader to the new file
	_, err = io.Copy(uncompressedFile, gzipReader)
	if err != nil {
		fmt.Printf("Error copying the file\n")
		return err
	}
	return nil
}

// Extract a tar archive into dest.
// For now this supports only tar archives, no compression.
func ExtractImage(archive, dest string) error {
	slog.Debug("ExtractImage", "archive", archive, "dest", dest)
	if !PathExists(dest) {
		os.MkdirAll(dest, 0777)
	}
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
