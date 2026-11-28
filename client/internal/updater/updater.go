package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var errNoUpdate = errors.New("已是最新版本，无需升级")

// Options 更新器运行参数（各入口注入）。
type Options struct {
	DataDir     string       // 设置与临时文件目录（BasePath/data）
	PanelURL    string       // 默认面板基址（连接地址），可被设置覆盖
	UID         string       // 灰度分桶标识（客户端/节点 key）
	Current     string       // 当前版本（common.VERSION）
	ServiceMode bool         // 服务模式：替换后交由服务管理器重启，不自行 relaunch+exit
	Notify      func(Info)   // notify 模式发现新版本回调（GUI 托盘/WebUI 徽标）
	Restart     func() error // 自定义重启（服务模式用 kardianos restart）；nil 则 relaunch+exit
	Log         func(string) // 日志回调
}

var (
	once     sync.Once
	mu       sync.Mutex
	opts     Options
	settings Settings
	pending  *Info
)

// Init 初始化/合并更新器参数与设置（不启动后台循环），供手动命令与 WebUI 按需调用。
// 采用合并语义：仅覆盖非空字段，保留已注入的 Notify/Restart 等回调（避免多入口互相清空）。
func Init(o Options) {
	if o.Current == "" {
		o.Current = "v0.0.0"
	}
	mu.Lock()
	if opts.DataDir == "" {
		opts = o
	} else {
		if o.DataDir != "" {
			opts.DataDir = o.DataDir
		}
		if o.PanelURL != "" {
			opts.PanelURL = o.PanelURL
		}
		if o.UID != "" {
			opts.UID = o.UID
		}
		if o.Current != "" {
			opts.Current = o.Current
		}
		if o.Log != nil {
			opts.Log = o.Log
		}
		if o.Notify != nil {
			opts.Notify = o.Notify
		}
		if o.Restart != nil {
			opts.Restart = o.Restart
		}
		opts.ServiceMode = opts.ServiceMode || o.ServiceMode
	}
	settings = LoadSettings(opts.DataDir)
	mu.Unlock()
	if exe, err := os.Executable(); err == nil {
		CleanupOld(exe)
	}
}

// Start 启动后台检查循环（sync.Once，多入口只启动一次）。Mode=off 时不启动。
func Start(o Options) {
	once.Do(func() {
		Init(o)
		mu.Lock()
		mode := settings.Mode
		mu.Unlock()
		if mode == ModeOff {
			return
		}
		go loop()
	})
}

func logf(format string, args ...interface{}) {
	if opts.Log != nil {
		opts.Log(fmt.Sprintf(format, args...))
	}
}

func loop() {
	interval := time.Duration(settings.IntervalHours) * time.Hour
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	checkOnce()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		mu.Lock()
		mode := settings.Mode
		hours := settings.IntervalHours
		mu.Unlock()
		if mode == ModeOff {
			continue
		}
		if d := time.Duration(hours) * time.Hour; d > 0 && d != interval {
			interval = d
			ticker.Reset(interval)
		}
		checkOnce()
	}
}

func baseURL() string {
	mu.Lock()
	defer mu.Unlock()
	if settings.PanelURL != "" {
		return settings.PanelURL
	}
	return opts.PanelURL
}

func checkOnce() {
	mu.Lock()
	channel := settings.Channel
	mode := settings.Mode
	mu.Unlock()
	info, err := CheckLatest(baseURL(), channel, opts.UID, opts.Current)
	if err != nil {
		logf("检查更新失败: %v", err)
		return
	}
	mu.Lock()
	pending = info
	mu.Unlock()
	if info == nil {
		return
	}
	logf("发现新版本 %s", info.Version)
	if mode == ModeAuto {
		if err := ApplyNow(); err != nil {
			logf("自动升级失败: %v", err)
		}
		return
	}
	if opts.Notify != nil {
		opts.Notify(*info)
	}
}

// CheckNow 手动检查更新（CLI/WebUI「检查更新」），返回新版本信息（无更新返回 nil）。
func CheckNow() (*Info, error) {
	mu.Lock()
	channel := settings.Channel
	mu.Unlock()
	info, err := CheckLatest(baseURL(), channel, opts.UID, opts.Current)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	pending = info
	mu.Unlock()
	return info, nil
}

// Pending 返回待升级版本（notify 模式徽标用）。
func Pending() *Info {
	mu.Lock()
	defer mu.Unlock()
	if pending == nil {
		return nil
	}
	cp := *pending
	return &cp
}

// GetSettings 返回当前设置。
func GetSettings() Settings {
	mu.Lock()
	defer mu.Unlock()
	return settings
}

// SaveSettings 保存设置（持久化 + 内存）。
func SaveSettings(s Settings) error {
	mu.Lock()
	defer mu.Unlock()
	if err := SaveSettingsFile(opts.DataDir, s); err != nil {
		return err
	}
	settings = s.Normalize()
	return nil
}

// ApplyNow 执行升级：下载→校验→解包→原子替换→重启。失败回滚。
func ApplyNow() error {
	mu.Lock()
	info := pending
	pubKey := settings.PublicKey
	autoRestart := settings.AutoRestart
	mu.Unlock()
	if info == nil {
		var err error
		info, err = CheckNow()
		if err != nil {
			return err
		}
		if info == nil {
			return errNoUpdate
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	tmpDir, err := os.MkdirTemp("", "gostc-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archive := filepath.Join(tmpDir, "update.archive")
	if err := Download(info.DownloadURL, archive); err != nil {
		return err
	}
	if err := VerifySha256(archive, info.Sha256); err != nil {
		return err
	}
	if err := VerifySignature(archive, info.Signature, pubKey); err != nil {
		return err
	}
	newBin := filepath.Join(tmpDir, "gostc-new")
	if runtime.GOOS == "windows" {
		newBin += ".exe"
	}
	if err := Extract(archive, newBin); err != nil {
		return err
	}
	if err := AtomicReplace(newBin, exe); err != nil {
		return err
	}
	logf("已替换二进制到 %s，版本 %s", exe, info.Version)
	mu.Lock()
	pending = nil
	mu.Unlock()

	if !autoRestart {
		return nil
	}
	if opts.Restart != nil {
		if err := opts.Restart(); err != nil {
			_ = Rollback(exe)
			return err
		}
		return nil
	}
	if opts.ServiceMode {
		return nil // 服务模式由外部管理器重启
	}
	if err := Relaunch(os.Args[1:]); err != nil {
		_ = Rollback(exe)
		return err
	}
	os.Exit(0)
	return nil
}
