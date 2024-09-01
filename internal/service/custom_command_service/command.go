package custom_command_service

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema/custom_command_schema"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
)

type cmdConfig struct {
	Commands []custom_command_schema.Command `yaml:"commands"`
}
type CustomCommandService struct {
	_conf *conf.Conf
}

func NewCustomCommandService(_conf *conf.Conf) *CustomCommandService {
	return &CustomCommandService{_conf: _conf}
}
func (u *CustomCommandService) ReadConfig(filePath string) (map[string]custom_command_schema.Command, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config cmdConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]custom_command_schema.Command, len(config.Commands))
	for _, command := range config.Commands {
		ret[command.Name] = command
	}

	return ret, nil
}
func (u *CustomCommandService) ExecuteCommand(cmd custom_command_schema.Command, stdout, stderr *custom_command_schema.CustomWriter) error {
	if u._conf.StartMode == conf.CommonMode || u._conf.StartMode == conf.DaemonMode {
		return u.executeCommand(cmd, stdout, stderr)
	}

	return nil
}
func (u *CustomCommandService) executeCommand(cmd custom_command_schema.Command, stdout, stderr *custom_command_schema.CustomWriter) error {
	command := exec.Command(cmd.Cmd, cmd.Args...)

	for key, value := range cmd.Env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, value))
	}
	command.Dir = cmd.WorkDir
	command.Stdout = stdout
	command.Stderr = stderr

	logger.Debugf("Executing command: %s", cmd.Name)

	if err := command.Start(); err != nil {
		logger.Warnf("Command %s failed with error: %v", cmd.Name, err)
		return err
	} else {
		logger.Warnf("Command %s executed successfully.", cmd.Name)
	}
	go func() {
		if err := command.Wait(); err != nil {

			logger.Warnf("Command %s failed with error: %v", cmd.Name, err)

		}
		stdout.Close()
		logger.Debugf("Command %s executed successfully.", cmd.Name)
	}()
	return nil

}
