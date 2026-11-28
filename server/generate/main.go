package main

import (
	"gorm.io/gen"
	"server/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./repository/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})

	g.ApplyBasic(
		model.GostAuth{},
		model.GostClient{},
		model.GostClientAdmission{},
		model.GostClientConfig{},
		model.GostClientForward{},
		model.GostClientHost{},
		model.GostClientHostDomain{},
		model.GostClientTunnel{},
		model.GostClientProxy{},
		model.GostClientP2P{},
		model.GostNode{},
		model.GostNodeBind{},
		model.GostNodeConfig{},
		model.GostNodeGroup{},
		model.GostFailoverLog{},
		model.GostNodeDomain{},
		model.GostNodePort{},
		model.GostObs{},
		model.SystemConfig{},
		model.SystemNotice{},
		model.SystemUser{},
		model.SystemUserGroup{},
		model.SystemUserCheckin{},
		model.SystemUserEmail{},
		model.SystemCdkBatch{},
		model.SystemCdk{},
		model.SystemAuditLog{},
		model.FrpClientCfg{},
		model.GostTunnelHealth{},
		model.SystemAlertRule{},
		model.SystemAlertLog{},
		model.GostMetric{},
		model.SystemCert{},
		model.SystemOrder{},
		model.SystemBill{},
		model.SystemApiToken{},
		model.SystemTunnelTemplate{},
		model.SystemClientVersion{},
	)
	g.Execute()
}
