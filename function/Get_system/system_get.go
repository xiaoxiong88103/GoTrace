package Get_system

import (
	"System_Log/function/Get_config"
	"System_Log/function/Get_proc"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// GPU的函数调用方法-------------------------------------------------------
// queryGPUInfo 用于查询GPU的特定信息
func queryGPUInfo(query string) ([]string, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu="+query, "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("err: %v", err)
	}

	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// formatGPUInfo 用于根据条件格式化GPU信息
func formatGPUInfo(temperatures, memoryUsed, memoryTotal, utilization []string, Temp, UsedMem, TotalMem, Per string) ([]string, error) {
	var formattedInfo []string
	for i := 0; i < len(temperatures); i++ {
		infoParts := []string{fmt.Sprintf("GPU%d:", i)}

		if Temp == "1" {
			infoParts = append(infoParts, fmt.Sprintf("Temp:%s°C", temperatures[i]))
		}
		if UsedMem == "1" {
			infoParts = append(infoParts, fmt.Sprintf("Used:%sMB", memoryUsed[i]))
		}
		if TotalMem == "1" {
			infoParts = append(infoParts, fmt.Sprintf("Total:%sMB", memoryTotal[i]))
		}
		if Per == "1" {
			infoParts = append(infoParts, fmt.Sprintf("Utilization:%s%%", utilization[i]))
		}

		formattedInfo = append(formattedInfo, strings.Join(infoParts, ", "))
	}

	return formattedInfo, nil
}

//GPU的函数调用方法-------------------------------------------------------

// System_get 返回系统的内存使用量（MB）和系统1分钟平均负载 获取CPU 内存 平均负载的
func System_get() (float64, uint64, float64, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, 0, 0.0, err
	}

	vmem, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0.0, err // 如果获取虚拟内存信息时出错，返回错误
	}
	mem := vmem.Used / 1024 / 1024 // 将内存使用量从字节转换为MB

	avg, err := load.Avg()
	if err != nil {
		return 0, 0, 0.0, err // 如果获取系统平均负载时出错，返回错误
	}

	return percent[0], mem, avg.Load1, nil // 返回内存使用量（MB）和系统1分钟平均负载
}

// 返回开机时间和当前时间（中国时区）
func Get_time() (louptime time.Time, currentTime time.Time, err error) {
	// 获取系统启动时间
	upTime, _ := host.Uptime()
	upSince := time.Now().Add(-time.Second * time.Duration(upTime))

	// 获取当前时间并转换为中国时区
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Time{}, time.Time{}, err // 如果加载时区失败，返回错误
	}
	currentTime = time.Now().In(loc)

	return upSince, currentTime, nil
}

// 返回进程的信息数据CPU 内存 运行时间 fd炳句 版本 pid号 错误
func Get_proc_info(pidname string) (float64, float32, time.Duration, int, string, int32, error) {
	pids := Get_proc.Get_PID(pidname)
	if len(pids) == 0 {
		fmt.Println(pidname, "process not found.")
		return 0, 0, 0, 0, "", 0, nil
	}

	pid := pids[0] // 取第一个PID

	// 获取进程信息
	memPercent, cpuPercent, runTime, handleCount, err := Get_proc.GetProcessInfo(pid)
	if err != nil {
		log.Fatalf("Error getting process info: %s\n", err)
	}
	mempercent := memPercent / 1024 / 1024
	// 读取版本信息
	verfile, err := Get_config.Get_config("pid", "verfile")
	if err != nil {
		verfile = ""
		fmt.Println("错误了读取版本:", err)
	}
	if verfile == "" {
		return cpuPercent, mempercent, runTime, handleCount, "", pid, err
	}
	version, err := Get_proc.ReadVersionFromFile(verfile)
	if err != nil {
		log.Fatalf("Error reading version file: %s\n", err)
	}
	return cpuPercent, mempercent, runTime, handleCount, version, pid, err
}

// 获取GPU的占用等信息
func GetGPULoad() []string {
	arch := runtime.GOARCH
	if arch != "amd64" {
		filePath := "/sys/devices/platform/fb000000.gpu/devfreq/fb000000.gpu/load"
		// 读取文件内容
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			// 读取文件失败时，返回错误信息的格式化字符串
			return []string{fmt.Sprintf("[Error: %v]", err)}
		}

		// 将文件内容转换为字符串
		loadInfo := string(content)

		// 使用字符串分割函数获取@前面的值
		parts := strings.Split(loadInfo, "@")
		if len(parts) != 2 {
			// 文件内容格式不符合预期时，返回一个错误信息
			return []string{"[Error: parsing GPU load]"}
		}

		// 提取@前面的值并返回
		loadValue := strings.TrimSpace(parts[0])

		return []string{fmt.Sprintf("[GPU:%s]", loadValue)}
	}

	// 分别查询GPU温度、已用内存、总内存和利用率
	temperatures, err := queryGPUInfo("temperature.gpu")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	memoryUsed, err := queryGPUInfo("memory.used")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	memoryTotal, err := queryGPUInfo("memory.total")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	utilization, err := queryGPUInfo("utilization.gpu")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	temp, err := Get_config.Get_config("gpu", "temp")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	umem, err := Get_config.Get_config("gpu", "umem")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}
	tmem, err := Get_config.Get_config("gpu", "tmem")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}
	per, err := Get_config.Get_config("gpu", "per")
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	// 格式化并输出GPU信息，根据控制变量决定显示哪些信息
	formattedInfo, err := formatGPUInfo(temperatures, memoryUsed, memoryTotal, utilization, temp, umem, tmem, per)
	if err != nil {
		return []string{fmt.Sprintf("[Error: %v]", err)}
	}

	return formattedInfo
}

// 获取NPU的
func GetNPULoad() string {
	// 使用 runtime.GOARCH 获取当前系统的架构
	arch := runtime.GOARCH

	// 检查当前系统的架构是否为 amd64
	if arch == "amd64" {
		return ""
	}

	filePath := "/sys/kernel/debug/rknpu/load"

	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("获取NPU负载信息时出错：%v\n", err)
		return ""
	}

	// 将文件内容转换为字符串
	loadInfo := string(content)

	// 使用字符串分割函数和字符串替换来格式化负载信息
	loadInfo = strings.ReplaceAll(loadInfo, "NPU load:", "")
	loadInfo = strings.TrimSpace(loadInfo)
	parts := strings.Split(loadInfo, ",")
	formattedLoad := []string{}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		formattedLoad = append(formattedLoad, part)
	}

	return strings.Join(formattedLoad, ",")
}

// 获取VPU的 GetMppServiceProcessID 获取使用 /dev/mpp_service 的进程ID或错误信息
func GetMppServiceProcessID() string {
	arch := runtime.GOARCH
	if arch != "amd64" {
		// 使用 exec.Command 执行 fuser 命令
		cmd := exec.Command("fuser", "/dev/mpp_service")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("执行 fuser 命令失败: %v", err)
		}

		// 将输出转换为字符串并分割以获取进程ID
		outputStr := strings.TrimSpace(string(output))
		parts := strings.Split(outputStr, ":")
		if len(parts) < 2 {
			return "解析 fuser 输出失败"
		}

		// 返回进程ID，去除可能的空格
		processID := strings.TrimSpace(parts[1])
		if processID == "" {
			return "没有进程使用 /dev/mpp_service"
		}
		return processID
	}
	return ""
}

// 获取剩余磁盘的大小 的GetDiskUsageInfo 返回包含所有磁盘分区的设备名、挂载点、剩余大小以及（可选的）占用百分比信息的字符串
func GetDiskUsageInfo() (string, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", fmt.Errorf("获取磁盘分区信息失败: %v", err)
	}

	var infos []string
	for _, partition := range partitions {
		usageStat, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // 如果获取某个分区信息失败，则跳过此分区
		}
		// 计算剩余大小（GB为单位）
		freeGB := float64(usageStat.Free) / (1024 * 1024 * 1024)
		info := fmt.Sprintf("磁盘: %s, 挂载点: %s, 剩余大小: %.2fGB", partition.Device, partition.Mountpoint, freeGB)

		// 如果showPercentage等于1，则添加占用百分比
		diskbfb, err := Get_config.Get_config("disk", "diskbfb")
		if err != nil {
			return "", fmt.Errorf("获取diskbfb失败: %v", err)
		}
		if diskbfb == "1" {
			usedPercentage := 100.0 - usageStat.UsedPercent
			info += fmt.Sprintf(", 占用百分比: %.2f%%", usedPercentage)
		}

		infos = append(infos, info)
	}

	// 使用" | "符号连接所有信息
	allInfo := "[ " + strings.Join(infos, " | ") + " ]"
	return allInfo, nil
}


// IO读写的使用率
func GetIOStats() []string {
	// 获取IO读写使用率
	ioStats, err := disk.IOCounters()
	if err != nil {
		fmt.Printf("Error getting IO counters: %v\n", err)
		return nil
	}

	var totalReadBytes, totalWriteBytes uint64
	for _, stat := range ioStats {
		totalReadBytes += stat.ReadBytes
		totalWriteBytes += stat.WriteBytes
	}

	// 将读写字节转换为MB/s
	ioReadUsage := float32(totalReadBytes) / 1024 / 1024
	ioWriteUsage := float32(totalWriteBytes) / 1024 / 1024

	// 休眠1秒钟，获取1秒内的IO使用率
	time.Sleep(1 * time.Second)

	// 再次获取IO读写使用率
	ioStats, err = disk.IOCounters()
	if err != nil {
		fmt.Printf("Error getting IO counters: %v\n", err)
		return nil
	}

	var newTotalReadBytes, newTotalWriteBytes uint64
	for _, stat := range ioStats {
		newTotalReadBytes += stat.ReadBytes
		newTotalWriteBytes += stat.WriteBytes
	}

	// 计算1秒内的读写字节增量
	readBytes := newTotalReadBytes - totalReadBytes
	writeBytes := newTotalWriteBytes - totalWriteBytes

	// 将增量转换为MB/s
	ioReadUsage = float32(readBytes) / 1024 / 1024
	ioWriteUsage = float32(writeBytes) / 1024 / 1024

	read, err := Get_config_int("io", "read")
	if err != nil {
		return []string{fmt.Sprintf("err: %v", err)}
	}

	write, err := Get_config_int("io", "write")
	if err != nil {
		return []string{fmt.Sprintf("err: %v", err)}
	}

	// 定义一个字符串切片，用于存储返回的结果
	var result []string

	// 如果read和write都为1，则返回包含两个值的字符串切片
	if read == 1 && write == 1 {
		result = []string{
			fmt.Sprintf("ioread: %.2f", ioReadUsage),
			fmt.Sprintf("iowrite: %.2f", ioWriteUsage),
		}
	} else if read == 1 {
		// 如果只有read为1，则返回只包含ioread的字符串切片
		result = []string{
			fmt.Sprintf("ioread: %.2f", ioReadUsage),
		}
	} else if write == 1 {
		// 如果只有write为1，则返回只包含iowrite的字符串切片
		result = []string{
			fmt.Sprintf("iowrite: %.2f", ioWriteUsage),
		}
	}

	return result
}

// 获取上下行和连接数的代码
func GetNetworkStats() []string {
	// 第一次获取网络IO统计
	lastStat, err := net.IOCounters(true)
	if err != nil {
		return []string{fmt.Sprintf("网卡部分失败err:", err)}
	}

	// 等待一秒
	time.Sleep(1 * time.Second)

	// 再次获取网络IO统计
	newStat, err := net.IOCounters(true)
	if err != nil {
		log.Printf("Error getting new network IO counters: %v", err)
		return []string{fmt.Sprintf("网卡部分失败err:", err)}
	}

	// 计算上传和下载速率
	var uploadRate, downloadRate float64
	for i := range newStat {
		if len(lastStat) > i { // 确保lastStat中有相应的索引
			recvDiff := float64(newStat[i].BytesRecv - lastStat[i].BytesRecv)
			transmitDiff := float64(newStat[i].BytesSent - lastStat[i].BytesSent)
			// 注意：这里修正了下载和上传速率的计算
			downloadRate += recvDiff / 1024 / 1024
			uploadRate += transmitDiff / 1024 / 1024
		}
	}

	// 获取网络连接数
	connections, err := net.Connections("all")
	if err != nil {
		return []string{fmt.Sprintf("网卡部分失败err:", err)}
	}

	up, err := Get_config_int("network", "up")
	down, err := Get_config_int("network", "down")
	nc, err := Get_config_int("network", "nc")

	var result []string

	if up == 1 {
		result = append(result, fmt.Sprintf("up: %.2f", uploadRate))
	}
	if down == 1 {
		result = append(result, fmt.Sprintf("down: %.2f", downloadRate))
	}
	if nc == 1 {
		result = append(result, fmt.Sprintf("nc: %d", len(connections)))
	}

	return result
}
