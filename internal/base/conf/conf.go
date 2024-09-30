package conf

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type StartMode uint8

const (
	Unknown StartMode = iota
	SlaveMode
	ServiceMode
	CommonMode
)

type Conf struct {
	workdir         string
	Debug           bool   `yaml:"debug"`
	LogName         string `yaml:"log_name"`
	LogLevel        string `yaml:"log_level"`
	EnableProfiling bool   `yaml:"enable_profiling"`
	StartMode       StartMode
	path            string
}

func (c *Conf) ReadConfigFromYml(filePath string) (string, error) {

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		return "", err
	}
	return filePath, nil
}
func (c *Conf) SetWorkdir(path string) {
	var err error
	c.workdir, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	c.workdir = path
}
func (c *Conf) GetWorkdir() string {
	return c.workdir
}
func (c *Conf) SetPath(path string) {
	c.path = path
}
