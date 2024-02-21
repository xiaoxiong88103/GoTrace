package Get_proc

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// Get_PID 示例，获取proc_name的PID
func Get_PID(pidname string) []int32 {
	processName := pidname
	pids, err := findPIDByName(processName)
	if err != nil {
		fmt.Printf("Error finding process: %s\n", err)
		return nil
	}
	if len(pids) == 0 {
		fmt.Printf("No process found with name %s\n", processName)
		return nil
	}
	return pids
}

// findPIDByName 查找具有指定名称的进程PID
func findPIDByName(processName string) ([]int32, error) {
	var pids []int32
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if name == processName {
			pids = append(pids, p.Pid)
		}
	}

	return pids, nil
}

// GetProcessInfo 获取特定PID进程的内存、CPU使用率和运行时间
func GetProcessInfo(pid int32) (float32, float64, time.Duration, int, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	memPercent, err := p.MemoryInfo()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	cpuPercent, err := p.CPUPercent()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	//将CPU转换成保留小数点前俩位数字
	formattedCPUPercent := fmt.Sprintf("%.2f", cpuPercent)

	// 将格式化后的字符串转换为 float64 类型
	cpuPercentFloat, err := strconv.ParseFloat(formattedCPUPercent, 64)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	createTime, err := p.CreateTime()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	fdPath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "fd")
	files, err := ioutil.ReadDir(fdPath)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// 计算运行时间
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	runTime := time.Duration(currentTime-int64(createTime)) * time.Millisecond

	return float32(memPercent.RSS), cpuPercentFloat, runTime, len(files), nil
}

// ReadVersionFromFile 读取并解析版本信息
func ReadVersionFromFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// 使用正则表达式匹配版本信息
	re := regexp.MustCompile(`Version: ([\d\.]+)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		// 未找到版本信息
		return "", fmt.Errorf("version information not found")
	}

	// 返回匹配的版本号
	return matches[1], nil
}
