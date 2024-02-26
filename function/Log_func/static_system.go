package Log_func

import (
	"System_Log/function/Get_config"
	"System_Log/function/Get_system"
	"bufio"
	"bytes"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// 如果没有创建目录
func ensureLogDir(dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			log.Fatalf("创建日志目录失败: %v", err)
		}
	}
}

// 收集系统信息并记录到日志 静态数据
func CollectAndLogSystemInfo(logDir string) {
	logFilePath := logDir + "/system_info.log"
	// 确保日志目录存在
	ensureLogDir(logDir)

	// 创建或打开日志文件
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}
	defer logFile.Close()

	// 初始化log包使用的logger
	logger := log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	// CPU信息
	arch := runtime.GOARCH
	if arch != "amd64" {
		cmd := exec.Command("lscpu")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			logger.Printf("执行lscpu命令失败: %v", err)
		}

		// 解析命令输出
		output := out.String()
		lines := strings.Split(output, "\n")
		info := make(map[string]string)
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				info[key] = value
			}
		}

		// 格式化所需的CPU信息
		cpuInfo := fmt.Sprintf("架构: %s, CPU型号: %s, 核心数: %s, 最大频率: %s MHz",
			info["Architecture"],
			info["Model name"],
			info["CPU(s)"],
			info["CPU max MHz"])
		logger.Printf(cpuInfo)

	} else {
		cpus, _ := cpu.Info()
		for _, cpu := range cpus {
			logger.Printf("CPU型号: %s, 核心数: %d, 频率: %.2fGHz\n", cpu.ModelName, cpu.Cores, cpu.Mhz/1000)
		}
	}

	// 内存信息
	vmStat, _ := mem.VirtualMemory()
	logger.Printf("内存总大小: %.2fGB\n", float64(vmStat.Total)/(1024*1024*1024))

	// 系统和内核版本
	hostInfo, _ := host.Info()
	logger.Printf("系统版本: %s, 内核版本: %s\n", hostInfo.PlatformVersion, hostInfo.KernelVersion)

	// 系统启动时间
	startTime := time.Unix(int64(hostInfo.BootTime), 0)
	logger.Printf("系统启动时间: %s\n", startTime.Format(time.RFC1123))

	//磁盘物理的大小信息
	allDiskInfo, err := GetAllDiskSizesAsString()
	if err != nil {
		logger.Printf("执行磁盘命令失败: %v", err)
	}
	logger.Printf("物理磁盘信息总大小: %s", allDiskInfo)

	// 磁盘挂载信息
	partitions, err := disk.Partitions(false)
	if err != nil {
		logger.Printf("获取磁盘分区信息失败: %v\n", err)
		return
	}
	//获取挂载磁盘的剩余大小
	for _, partition := range partitions {
		usageStat, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			logger.Printf("获取磁盘使用情况失败: %v\n", err)
			continue // 如果获取某个分区信息失败，则跳过此分区
		}
		// 计算总大小和剩余大小（GB为单位）
		totalGB := float64(usageStat.Total) / (1024 * 1024 * 1024)
		freeGB := float64(usageStat.Free) / (1024 * 1024 * 1024)

		logger.Printf("磁盘: %s, 挂载点: %s, 总大小: %.2fGB, 剩余大小: %.2fGB\n", partition.Device, partition.Mountpoint, totalGB, freeGB)
	}

	// 注意:VPU信息的收集可能需要特定的方法
	gpu, err := Get_config.Get_config_int("static", "gpu")
	if err != nil {
		fmt.Println("无法读取gpu部分:", err)
		return
	}
	npu, err := Get_config.Get_config_int("static", "npu")
	if err != nil {
		fmt.Println("无法读取npu部分:", err)
		return
	}
	vpu, err := Get_config.Get_config_int("static", "npu")
	if err != nil {
		fmt.Println("无法读取vpu部分:", err)
		return
	}
	if gpu == 1 {
		logger.Printf("GPU信息: %s", Get_system.GetGPULoad())
		logger.Printf("GPU参数: %s", getGPUInfo())
	} else if npu == 1 {
		logger.Printf("NPU信息: %s", Get_system.GetNPULoad())
	} else if vpu == 1 {
		logger.Printf("VPU被当前进程调用中: %s", Get_system.GetMppServiceProcessID())
	}

	//来删除多余的行数
	TrimLogFile(logFilePath)

}

// GetAllDiskSizesAsString 使用fdisk -l命令自动获取所有磁盘的大小信息，并以一个大字符串返回
func GetAllDiskSizesAsString() (string, error) {
	cmd := exec.Command("fdisk", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("执行fdisk命令失败: %v", err)
	}

	re := regexp.MustCompile(`Disk (/dev/\w+): ([\d.]+\s+[TGM]i?B)`)
	matches := re.FindAllStringSubmatch(out.String(), -1)
	if matches == nil {
		return "", fmt.Errorf("未找到任何磁盘的大小信息")
	}

	var diskInfos []string
	for _, match := range matches {
		diskInfo := fmt.Sprintf("%s 磁盘大小%s", match[1], match[2])
		diskInfos = append(diskInfos, diskInfo)
	}

	// 使用特定符号连接所有磁盘信息
	allDiskInfo := "[" + strings.Join(diskInfos, " ｜") + "]"
	return allDiskInfo, nil
}

// TrimLogFile 检查日志文件的行数，如果超出限制，则从头部删除多余的行
func TrimLogFile(filePath string) error {
	maxLines, err := Get_config.Get_config_int("log", "logline")
	if err != nil {
		return err
	}
	// 打开日志文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取文件内容
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// 检查是否需要删除行
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:] // 保留最新的maxLines行
	}

	// 将调整后的内容写回文件
	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

// 获取GPU参数信息
func getGPUInfo() string {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,driver_version", "--format=csv,noheader")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("执行命令时出错:", err)
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpuList []string
	for i, line := range lines {
		gpuList = append(gpuList, fmt.Sprintf("%d:%s", i, strings.TrimSpace(line)))
	}

	return fmt.Sprintf("[%s | number:%d]", strings.Join(gpuList, " | "), len(lines))
}
