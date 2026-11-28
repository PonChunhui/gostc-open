package global

import (
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gostc-sub/pkg/fs"
	"sync"
)

var Logger *zap.Logger

var Cron *cron.Cron

// DataDir 数据目录（BasePath/data），自更新设置存储于此。
var DataDir string

var ClientFS *fs.FileStorage
var TunnelFS *fs.FileStorage
var P2PFS *fs.FileStorage

var ClientMap = sync.Map{}
var P2PMap = sync.Map{}
var TunnelMap = sync.Map{}
