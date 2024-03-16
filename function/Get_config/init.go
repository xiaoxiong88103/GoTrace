package Get_config

import (
	"fmt"
	"os"
	"path/filepath"
)

var Path_Config string

func init() {
	Path_Config = Dirfile("./config/")
}

func Dirfile(dirname string) string {
	// 获取当前执行文件的完整路径
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return "nil"
	}

	// 获取执行文件所在的目录
	execDir := filepath.Dir(execPath)

	// 构建相对于执行文件位置的路径
	relativePath := filepath.Join(execDir, dirname)
	addxiegang := relativePath + "/"
	return addxiegang
}
