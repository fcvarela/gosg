// Package filesystem provides a core.ResourceSystem which loads resources from the filesystem
package filesystem

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fcvarela/gosg/core"
	"github.com/golang/glog"
)

// ResourceSystem implements the resource system interface
type ResourceSystem struct {
	paths map[string][]string
}

var (
	basePath     = flag.String("data", "./data", "Data directory")
	userBasePath = flag.String("appdata", "./appdata", "User application data directory")
)

func init() {
	flag.Parse()
	core.GetResourceManager().SetSystem(New())
}

// New returns a new ResourceSystem
func New() *ResourceSystem {
	var bp, ubp string

	if runtime.GOOS == "darwin" && strings.HasSuffix(filepath.Dir(os.Args[0]), "MacOS") {
		glog.Info("Looking for data directory in same folder")
		path, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/../Resources")
		if err != nil {
			glog.Fatalf("Could not create data path from provided: %s\n", *basePath)
		}
		bp = filepath.Join(path, *basePath)
		ubp = filepath.Join(path, *userBasePath)
	} else {
		path, err := filepath.Abs(*basePath)
		if err != nil {
			glog.Fatalf("Could not create data path from provided: %s\n", *basePath)
		}

		userPath, err := filepath.Abs(*userBasePath)
		if err != nil {
			glog.Fatalf("Could not create data path from provided: %s\n", *basePath)
		}

		bp = path
		ubp = userPath
	}

	paths := make(map[string][]string)
	paths["programs"] = []string{filepath.Join(bp, "programs"), filepath.Join(ubp, "programs")}
	paths["states"] = []string{filepath.Join(bp, "states"), filepath.Join(ubp, "states")}
	paths["models"] = []string{filepath.Join(bp, "models"), filepath.Join(ubp, "models")}
	paths["textures"] = []string{filepath.Join(bp, "textures"), filepath.Join(ubp, "textures")}

	r := ResourceSystem{paths: paths}

	for _, p := range paths {
		if _, err := os.Stat(p[0]); os.IsNotExist(err) {
			glog.Fatalf("No such file or directory: %v\n", p[0])
		}
	}

	return &r
}

func (r *ResourceSystem) resourceWithFullpath(fullpath string) ([]byte, error) {
	data, e := ioutil.ReadFile(fullpath)
	if e != nil {
		return nil, fmt.Errorf("could not read file: %v", e)
	}
	return data, nil
}

func (r *ResourceSystem) loadResource(name, rtype string) []byte {
	fullpath := filepath.Join(r.paths[rtype][1], name)
	res, err := r.resourceWithFullpath(fullpath)
	if err != nil {
		fullpath = filepath.Join(r.paths[rtype][0], name)
		res, err = r.resourceWithFullpath(fullpath)
		if err != nil {
			glog.Fatal(err)
		}
		return res
	}
	return res
}

// Model implements the core.ResourceSystem interface
func (r *ResourceSystem) Model(filename string) []byte {
	return r.loadResource(filename, "models")
}

// Texture implements the core.ResourceSystem interface
func (r *ResourceSystem) Texture(filename string) []byte {
	return r.loadResource(filename, "textures")
}

// State implements the core.ResourceSystem interface
func (r *ResourceSystem) State(name string) []byte {
	var filename = name + ".json"
	return r.loadResource(filename, "states")
}

// Program implements the core.ResourceSystem interface
func (r *ResourceSystem) Program(name string) []byte {
	var filename = name + "." + core.GetRenderSystem().ProgramExtension()
	return r.loadResource(filename, "programs")
}

// ProgramData implements the core.ResourceSystem interface
func (r *ResourceSystem) ProgramData(name string) []byte {
	return r.loadResource(name, "programs")
}
