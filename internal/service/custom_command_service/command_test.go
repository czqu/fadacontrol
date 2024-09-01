package custom_command_service

import (
	"fadacontrol/internal/schema/custom_command_schema"
	"testing"
)

func TestExecuteCommand(t *testing.T) {
	// 创建 CustomCommandService 实例
	service := NewCustomCommandService()

	// 定义测试用例
	testCases := []struct {
		name string
		cmd  custom_command_schema.Command
		err  bool
	}{
		{
			name: "SuccessfulExecution",
			cmd: custom_command_schema.Command{
				Name: "test_dir",
				Cmd:  "cmd",
				Args: []string{"/C", "dir"},
				Env:  map[string]string{},
			},
			err: false,
		},
		{
			name: "SuccessfulExecution",
			cmd: custom_command_schema.Command{
				Name: "test_ls",
				Cmd:  "powershell",
				Args: []string{"-Command", "ls"},
				Env:  map[string]string{},
			},
			err: false,
		},
		{
			name: "SuccessfulExecution",
			cmd: custom_command_schema.Command{
				Name: "test_env",
				Cmd:  "powershell",
				Args: []string{"-Command", "$env:PATH"},
				Env:  map[string]string{},
			},
			err: false,
		},
		{
			name: "SuccessfulExecution",
			cmd: custom_command_schema.Command{
				Name: "test_env",
				Cmd:  "powershell",
				Args: []string{"-Command", "$env:RFU"},
				Env:  map[string]string{"RFU": "123"},
			},
			err: false,
		},
	}

}
