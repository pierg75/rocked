package cmd

import (
	"log"
	"os"
	"strconv"

	"log/slog"
)

var (
	BASE_CG_PATH        = "/sys/fs/cgroup/rocked/"
	BASE_CG_CONTROLLERS = []string{"cpu", "io", "memory", "pids"}
)

type Cgroup struct {
	Id            string
	path          string
	CgroupConPath string
	CpuLimit      string
	MemLimit      int
}

func NewCgroup(id string) *Cgroup {
	slog.Debug("Cgroup: Initialising new cgroup", "id", id)
	return &Cgroup{
		Id:            id,
		path:          BASE_CG_PATH,
		CgroupConPath: BASE_CG_PATH + id,
		CpuLimit:      "200000 1000000",
		MemLimit:      1 * 1024 * 1024 * 1024, // 1GB
	}
}

func (c *Cgroup) Path() string {
	return c.path
}

func (c *Cgroup) SetPath(path string) {
	c.path = path
}

func (c *Cgroup) CreateDirs() error {
	return os.MkdirAll(c.path, 0770)
}

// Make sure the subtrees can use the cpu, io, memory and pids controllers
func (c *Cgroup) SetControllers() error {
	controlPath := c.path + "cgroup.subtree_control"
	ctrlf, err := os.OpenFile(controlPath, os.O_RDWR, 0644)
	if err != nil {
		slog.Debug("Cgroup SetControllers error opening the subtree_control file", "err", err)
		return err
	}
	defer ctrlf.Close()
	for _, ctrl := range BASE_CG_CONTROLLERS {
		_, err = ctrlf.Write([]byte("+" + ctrl))
		if err != nil {
			log.Fatal("Error writing the controller ", controlPath, ": ", err)
			slog.Debug("Cgroup SetControllers error writing to the subtree_control file", "ctrl", ctrl, "err", err)
			return err
		}
	}
	return nil
}

// Creates the container cgroup directory
func (c *Cgroup) CreateConCgroup() error {
	slog.Debug("Cgroup CreateConCgroup", "CgroupConPath", c.CgroupConPath)
	err := os.MkdirAll(c.CgroupConPath, 0770)
	if err != nil {
		slog.Debug("Cgroup CreateConCgroup error creating the container Cgroup dir", "CgroupConPath", c.CgroupConPath, "err", err)
		return err
	}
	return nil
}

// Return the file reference to be later used with the clone3 syscall
func (c *Cgroup) GetCGFd() (*os.File, error) {
	cgroupControlFile, err := os.Open(c.CgroupConPath)
	if err != nil {
		slog.Debug("Cgroup GetCGFd error opening Cgroup dir", "baseContainerCgroupPath", c.CgroupConPath, "err", err)
		return nil, err
	}
	return cgroupControlFile, nil
}

// Sets container limits. For now this is limited to the cpu and memory
func (c *Cgroup) SetCGLimits() error {
	err := c.setCgroupMaxLimit("cpu", c.CpuLimit)
	if err != nil {
		return err
	}
	err = c.setCgroupMaxLimit("memory", strconv.Itoa(c.MemLimit))
	if err != nil {
		return err
	}
	return nil
}

// Set a controller max setting.
// Note that some controllers take a single parameter while some take more
func (c *Cgroup) setCgroupMaxLimit(controller, setting string) error {
	ctrlMax, err := os.OpenFile(c.CgroupConPath+"/"+controller+".max", os.O_RDWR, 0644)
	if err != nil {
		slog.Debug("Cgroup SetCGLimits error opening", "controller", controller, "CgroupConPath", c.CgroupConPath, "err", err)
		return err
	}
	defer ctrlMax.Close()
	_, err = ctrlMax.Write([]byte(setting))
	if err != nil {
		slog.Debug("Cgroup SetCGLimits error writing", "controller", controller, "CgroupConPath", c.CgroupConPath, "err", err)
		return err
	}
	return nil
}

// Create a new cgroup, sets the controllers and create the necessary directories.
func PrepareCgroup(con *Container, cArgs *CloneArgs) (*Cgroup, error) {
	slog.Debug("prepareCgroup", "ID", con.id, "cArgs", cArgs)
	cg := NewCgroup(con.id)
	err := cg.SetControllers()
	if err != nil {
		return nil, err
	}
	err = cg.CreateConCgroup()
	if err != nil {
		return nil, err
	}
	slog.Debug("prepareCgroup", "returning", nil)
	return cg, nil
}
