package api

import (
	"net"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// 用于返回给前端的客户端信息
type ClientInfo struct {
	UUID       string `json:"uuid"`
	RemoteAddr string `json:"remote_addr"`
	LoginTime  int64  `json:"login_time"`
	Heartbeat  int64  `json:"heartbeat"`
	ConnTime   int64  `json:"conn_time"`
}

func GetClients(c *gin.Context) {

}

// ServerInfo 服务器信息结构体
type ServerInfo struct {
	GoroutineCount int              `json:"goroutine_count"` // goroutine数量
	MemoryStats    runtime.MemStats `json:"memory_stats"`    // 内存统计信息
	NetworkInfo    NetworkInfo      `json:"network_info"`    // 网络信息
	Uptime         int64            `json:"uptime"`          // 运行时间(秒)
	ClientCount    int              `json:"client_count"`    // 当前连接的客户端数量
}

// NetworkInfo 网络信息结构体
type NetworkInfo struct {
	Interfaces []InterfaceInfo `json:"interfaces"` // 网络接口信息
}

// InterfaceInfo 网络接口信息
type InterfaceInfo struct {
	Name      string   `json:"name"`      // 接口名称
	Addresses []string `json:"addresses"` // IP地址列表
	Flags     string   `json:"flags"`     // 接口标志
}

var serverStartTime = time.Now() // 服务器启动时间

func GetServerInfor(c *gin.Context) {

}

// getNetworkInfo 获取网络接口信息
func getNetworkInfo() NetworkInfo {
	var networkInfo NetworkInfo

	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return networkInfo
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取接口地址
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var addresses []string
		for _, addr := range addrs {
			// 只获取IP地址，排除网络地址
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil {
					addresses = append(addresses, ipNet.IP.String())
				}
			}
		}

		if len(addresses) > 0 {
			interfaceInfo := InterfaceInfo{
				Name:      iface.Name,
				Addresses: addresses,
				Flags:     iface.Flags.String(),
			}
			networkInfo.Interfaces = append(networkInfo.Interfaces, interfaceInfo)
		}
	}

	return networkInfo
}
