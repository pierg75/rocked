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

var (
	VALID_ALGO = []string{"sha256", "sha512"}
)

type MediaTypeError struct{}

func (m *MediaTypeError) Error() string {
	return "Wrong MediaType"
}

type AlgorithmError struct{}

func (m *AlgorithmError) Error() string {
	return "Wrong Algorithm"
}

func IsValidAlgorithm(algo string) bool {
	for _, v := range VALID_ALGO {
		if v == algo {
			return true
		}
	}
	return false
}

type Container struct {
	Path          string
	JsonPath      string
	Index         specs.Index
	ImageManifest specs.Manifest
	Image         specs.Image
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

func (c *Container) ReadImageManifest(manifest specs.Descriptor) error {
	slog.Debug("Container: ReadImageManifest", "manifest", manifest)
	if manifest.MediaType != "application/vnd.oci.image.manifest.v1+json" {
		return &MediaTypeError{}
	}
	algo := string(manifest.Digest.Algorithm())
	if !IsValidAlgorithm(algo) {
		return &AlgorithmError{}
	}
	imagePath := c.Path + "/blobs/" + algo + "/" + string(manifest.Digest.Encoded())
	slog.Debug("Container: ReadImageManifest", "imagePath", imagePath)
	jsonblob, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer jsonblob.Close()
	byteValue, _ := io.ReadAll(jsonblob)
	var image specs.Manifest
	err = json.Unmarshal(byteValue, &image)
	if err != nil {
		return err
	}
	slog.Debug("Container: ", "image", image)
	c.ImageManifest = image
	return nil
}

func (c *Container) ReadImageConfig() error {
	config := c.ImageManifest.Config
	slog.Debug("Container: ReadImageConfig", "config", config)
	if config.MediaType != "application/vnd.oci.image.config.v1+json" {
		return &MediaTypeError{}
	}
	algo := string(config.Digest.Algorithm())
	if !IsValidAlgorithm(algo) {
		return &AlgorithmError{}
	}
	configPath := c.Path + "/blobs/" + algo + "/" + string(config.Digest.Encoded())
	slog.Debug("Container: ReadImageConfig", "configPath", configPath)
	jsonblob, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer jsonblob.Close()
	byteValue, _ := io.ReadAll(jsonblob)
	var configjson specs.Image
	err = json.Unmarshal(byteValue, &configjson)
	if err != nil {
		return err
	}
	slog.Debug("Container: ", "config", configjson)
	c.Image = configjson
	return nil
}
func (c *Container) GetDigestPath() {

}
