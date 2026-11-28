package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gostc-sub/internal/common"
	"gostc-sub/internal/updater"
	"gostc-sub/pkg/env"

	system_service "github.com/kardianos/service"
)

// cmdUpdate 处理 `gostc update` 命令：检查新版本，--yes 时下载校验替换并重启。
// 命令行版本与服务模式均支持；服务模式升级后通过服务管理器重启。
func cmdUpdate(args []string) {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	var address string
	fs.StringVar(&address, "addr", "gostc.shop", "server address")
	var tlsEnable bool
	fs.BoolVar(&tlsEnable, "tls", true, "enable tls")
	var key string
	fs.StringVar(&key, "key", "", "client/node key（灰度分桶）")
	var mode string
	fs.StringVar(&mode, "mode", "", "更新模式 off|notify|auto（覆盖设置）")
	var channel string
	fs.StringVar(&channel, "channel", "", "更新通道 stable|beta（覆盖设置）")
	var interval int
	fs.IntVar(&interval, "interval", 0, "检查间隔小时（覆盖设置）")
	var panelURL string
	fs.StringVar(&panelURL, "url", "", "面板基址覆盖")
	var yes bool
	fs.BoolVar(&yes, "yes", false, "发现新版本时立即升级")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	address = env.GetEnv("GOSTC_CLIENT_ADDR", address)
	tlsEnable = env.GetEnv("GOSTC_CLIENT_TLS", tlsEnable)
	mode = env.GetEnv("GOSTC_UPDATE_MODE", mode)
	channel = env.GetEnv("GOSTC_UPDATE_CHANNEL", channel)
	interval = env.GetEnv("GOSTC_UPDATE_INTERVAL", interval)
	panelURL = env.GetEnv("GOSTC_UPDATE_URL", panelURL)

	exe, _ := os.Executable()
	dataDir := filepath.Join(filepath.Dir(exe), "data")

	// 应用旗标/env 覆盖并持久化
	s := updater.LoadSettings(dataDir)
	if mode != "" {
		s.Mode = mode
	}
	if channel != "" {
		s.Channel = channel
	}
	if interval > 0 {
		s.IntervalHours = interval
	}
	if panelURL != "" {
		s.PanelURL = panelURL
	}
	_ = updater.SaveSettingsFile(dataDir, s)

	generate := common.NewGenerateUrl(tlsEnable, address)
	base := generate.HttpUrl()
	if s.PanelURL != "" {
		base = s.PanelURL
	}

	interactive := system_service.Interactive()
	updater.Init(updater.Options{
		DataDir:     dataDir,
		PanelURL:    base,
		UID:         key,
		Current:     common.VERSION,
		ServiceMode: !interactive,
		Restart: func() error {
			svr, err := system_service.New(Program, SvcCfg)
			if err != nil {
				return err
			}
			return svr.Restart()
		},
		Log: func(msg string) { fmt.Println(msg) },
	})

	fmt.Println("当前版本：", common.VERSION)
	info, err := updater.CheckNow()
	if err != nil {
		fmt.Println("检查更新失败：", err)
		os.Exit(1)
	}
	if info == nil {
		fmt.Println("已是最新版本，无需升级。")
		return
	}
	fmt.Printf("发现新版本：%s（通道 %s）\n", info.Version, info.Channel)
	if info.ReleaseNote != "" {
		fmt.Println("更新说明：", info.ReleaseNote)
	}
	if !yes {
		fmt.Println("执行升级请运行：gostc update --yes")
		return
	}
	fmt.Println("开始升级……")
	if err := updater.ApplyNow(); err != nil {
		fmt.Println("升级失败：", err)
		os.Exit(1)
	}
	fmt.Println("升级完成。")
}
