package model

// SystemClientVersion 客户端版本发布（自动更新）。
type SystemClientVersion struct {
	Base
	Version     string `gorm:"column:version;index;comment:版本号"`
	Platform    string `gorm:"column:platform;index;comment:平台windows/linux/darwin"`
	Arch        string `gorm:"column:arch;comment:架构amd64/arm64"`
	DownloadURL string `gorm:"column:download_url;comment:下载地址"`
	Sha256      string `gorm:"column:sha256;comment:包SHA256"`
	Signature   string `gorm:"column:signature;comment:包签名"`
	Channel     string `gorm:"column:channel;default:'stable';comment:通道stable/beta"`
	GrayPercent int    `gorm:"column:gray_percent;default:100;comment:灰度百分比0-100"`
	Enable      int    `gorm:"column:enable;size:1;default:1;comment:启用"`
	ReleaseNote string `gorm:"column:release_note;size:1000;comment:更新说明"`
}

const (
	CLIENT_VERSION_CHANNEL_STABLE = "stable"
	CLIENT_VERSION_CHANNEL_BETA   = "beta"
)
