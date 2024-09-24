package conf

import (
	"fmt"
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
	workdir   string
	Debug     bool   `yaml:"debug"`
	LogName   string `yaml:"log_name"`
	LogLevel  string `yaml:"log_level"`
	StartMode StartMode
}

func (c *Conf) ReadConfigFromYml(filePath string) error {

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
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
