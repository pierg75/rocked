package cmd

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"rocked/specs"

	"log/slog"

	"github.com/opencontainers/go-digest"
)

type Container struct {
	Path     string
	JsonPath string
	Index    specs.Index
}

func NewContainer(path string) *Container {
	jpath := path + "/index.json"
	slog.Debug("Container: Initialising new container", "path", path, "JsonPath", jpath)
	if len(path) == 0 {
		log.Panic("Path cannot be empty")
	}
	return &Container{
		Path:     path,
		JsonPath: jpath,
	}
}

func (c *Container) LoadConfigJson() error {
	slog.Debug("Container: Loading index")
	jsonconfig, err := os.Open(c.JsonPath)

	if err != nil {
		return err
	}
	defer jsonconfig.Close()
	byteValue, _ := io.ReadAll(jsonconfig)
	var index specs.Index
	err = json.Unmarshal(byteValue, &index)
	if err != nil {
		return err
	}
	slog.Debug("Container: ", "index", index)
	c.Index = index
	return nil
}

func (c *Container) BlobExists(algo digest.Algorithm, digest string) bool {
	digestPath := c.Path + "/blobs/" + algo.String() + "/" + digest
	slog.Debug("Container: BlobExists", "digest path", digestPath)
	_, err := os.Stat(digestPath)
	if os.IsNotExist(err) {
		slog.Debug("Container: BlobExists", "return", false)
		return false
	}
	slog.Debug("Container: BlobExists", "return", true)
	return true
}
