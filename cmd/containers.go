package cmd

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"rocked/specs"
	"rocked/utils"

	"log/slog"

	"github.com/google/uuid"
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
	id            string
	Index         specs.Index
	ImageManifest specs.Manifest
	Image         specs.Image
}

func NewContainer(path string) *Container {
	slog.Debug("Container: Initialising new container", "path", path)
	if len(path) == 0 {
		log.Panic("Path cannot be empty")
	}
	id := uuid.New().String()
	cpath := path + id
	jpath := cpath + "/index.json"
	slog.Debug("Container: Initialising new container", "cpath", cpath, "JsonPath", jpath)
	return &Container{
		id:       id,
		Path:     cpath,
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

func (c *Container) ExpandAllManifest(defaultContainerImage string) error {
	slog.Debug("ExpandAllManifest", "defaultContainerImage", defaultContainerImage)
	for _, manifest := range c.Index.Manifests {
		exists := c.BlobExists(manifest.Digest.Algorithm(), manifest.Digest.Encoded())
		if exists {
			slog.Debug("setContainer", "manifest", manifest.Annotations["org.opencontainers.image.ref.name"], "exists", exists)
			err := c.ReadImageManifest(manifest)
			if err != nil {
				return err
			}
			err = c.ReadImageConfig()
			if err != nil {
				return err
			}

			for _, layer := range c.ImageManifest.Layers {
				// Extract the base layer into defaultContainerImage+"image_root"
				layerPath := defaultContainerImage + "/blobs/" + string(layer.Digest.Algorithm()) + "/" + string(layer.Digest.Encoded())
				slog.Debug("setContainer", "layer path", layerPath)
				//utils.Gunzip(layerPath, "/home/plambri/tempunzipped")
				containerImageRoot := defaultContainerImage + "/image_root"
				if !utils.PathExists(containerImageRoot) {
					os.Mkdir(containerImageRoot, 0770)
				}
				cmd := exec.Command("tar", "xvf", layerPath, "-C", containerImageRoot)
				if err := cmd.Run(); err != nil {
					log.Printf("Error while unpacking the layer %v: error %v", layerPath, err)
					return err
				}
			}
		}
	}
	return nil
}

func (c *Container) GetDigestPath() {}

// Create the necessary directories (work, upper and merge) in <path>
func CreateOverlayDirs(path string) error {
	slog.Debug("createOverlayDirs", "path", path)
	for _, dir := range []string{"work", "upper", "merge"} {
		err := os.MkdirAll(path+"/overlay/"+dir, 0770)
		if err != nil {
			slog.Debug("createOverlayDirs", "dir", dir, "error", err)
			return err
		}
	}
	return nil
}

func SetContainer(image, base_path string) (*Container, error) {
	slog.Debug("setContainert", "image", image, "base_path", base_path)
	con := NewContainer(base_path)
	var defaultSourceContainersImages string = "blobs/container_images/" + image + ".tar"
	utils.ExtractImage(defaultSourceContainersImages, con.Path)
	errcon := con.LoadConfigJson()
	if errcon != nil {
		return nil, errcon
	}
	slog.Debug("setContainert", "Manifests", con.Index.Manifests)
	con.ExpandAllManifest(con.Path)
	err := CreateOverlayDirs(con.Path)
	if err != nil {
		return nil, err
	}
	return con, nil
}
