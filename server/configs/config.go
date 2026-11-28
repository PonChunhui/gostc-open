package configs

import (
	"time"
)

type Config struct {
	Address   string        `yaml:"address"`
	AuthKey   string        `yaml:"auth-key"`
	AuthExp   time.Duration `yaml:"auth-exp"`
	AuthRenew time.Duration `yaml:"auth-renew"`
	DbType    string        `yaml:"db-type"`
	Sqlite    Sqlite        `yaml:"sqlite"`
	Mysql     Mysql         `yaml:"mysql"`
	Cache     Cache         `yaml:"cache"`
	Encrypt   Encrypt       `yaml:"encrypt"`
	Health    Health        `yaml:"health"`
	Metric    Metric        `yaml:"metric"`
	ACME      ACME          `yaml:"acme"`
	DBResolver DBResolver   `yaml:"db-resolver"`
	Pay       Pay           `yaml:"pay"`
	DNS       DNS           `yaml:"dns"`
	Failover  Failover      `yaml:"failover"`
}
