package global

// BasePath 为程序数据根目录（其下再建 data 子目录存放配置）。
// Windows：可执行文件所在目录；macOS：~/Library/Application Support/GOSTC
// （.app 包内不可写且受 Gatekeeper 路径随机化影响，必须放到用户目录）。
var BasePath string

func init() {
	BasePath = defaultBasePath()
}
