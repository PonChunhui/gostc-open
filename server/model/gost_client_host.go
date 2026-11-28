package model

import "fmt"

type GostClientHost struct {
	Base
	Name                string     `gorm:"column:name;index;comment:名称"`
	TargetIp            string     `gorm:"column:target_ip;index;comment:内网IP"`
	TargetPort          string     `gorm:"column:target_port;index;comment:内网端口"`
	TargetHttps         int        `gorm:"column:target_https;comment:是否为Https服务"`
	DomainPrefix        string     `gorm:"column:domain_prefix;index;comment:域名前缀"`
	CustomDomain        string     `gorm:"column:custom_domain;index;comment:自定义域名"`
	CustomCert          string     `gorm:"column:custom_cert;comment:自定义证书"`
	CustomKey           string     `gorm:"column:custom_key;comment:自定义证书"`
	CustomForceHttps    int        `gorm:"column:custom_force_https;size:1;comment:强制HTTPS"`
	CustomDomainMatcher int        `gorm:"column:custom_domain_matcher;size:1;comment:是否为泛域名"`
	NodeCode            string     `gorm:"column:node_code;index;comment:节点编号"`
	BackupNodeCode      string     `gorm:"column:backup_node_code;index;comment:备用节点编号"`
	FailoverEnable      int        `gorm:"column:failover_enable;size:1;default:2;comment:故障转移"`
	Node                GostNode   `gorm:"foreignKey:NodeCode;references:Code"`
	ClientCode          string     `gorm:"column:client_code;index;comment:客户端编号"`
	Client              GostClient `gorm:"foreignKey:ClientCode;references:Code"`
	UserCode            string     `gorm:"column:user_code;index;comment:用户编号"`
	User                SystemUser `gorm:"foreignKey:UserCode;references:Code"`
	Enable              int        `gorm:"column:enable;size:1;default:1;comment:启用状态"`
	Status              int        `gorm:"column:status;size:1;default:1;comment:状态"`
	BanReason           string     `gorm:"column:ban_reason;size:500;default:'';not null;comment:封禁原因"`
	UseEncryption       int        `gorm:"column:use_encryption;size:1;default:1;comment:加密"`
	UseCompression      int        `gorm:"column:use_compression;size:1;default:1;comment:压缩"`
	PoolCount           int        `gorm:"column:pool_count;default:0;comment:复用连接数量"`
	Locations           string     `gorm:"column:locations;comment:HTTP路由locations(JSON数组)"`
	HostHeaderRewrite   string     `gorm:"column:host_header_rewrite;comment:回源Host改写"`
	LoadBalancerGroup   string     `gorm:"column:load_balancer_group;comment:负载均衡组"`
	LoadBalancerGroupKey string    `gorm:"column:load_balancer_group_key;comment:负载均衡组密钥"`
	HealthCheckType     string     `gorm:"column:health_check_type;comment:健康检查类型tcp/http"`
	HealthCheckPath     string     `gorm:"column:health_check_path;comment:健康检查路径"`
	HealthCheckInterval int        `gorm:"column:health_check_interval;default:0;comment:健康检查间隔秒"`
	HealthCheckTimeout  int        `gorm:"column:health_check_timeout;default:0;comment:健康检查超时秒"`
	HealthCheckMaxFailed int       `gorm:"column:health_check_max_failed;default:0;comment:健康检查最大失败次数"`
	GostClientAdmission
	GostClientConfig
}

func (host *GostClientHost) GetLocations() []string {
	return parseStringList(host.Locations)
}

func (host *GostClientHost) SetLocations(locations []string) {
	host.Locations = MarshalStringList(locations)
}

func (host *GostClientHost) GetCustomDomain() string {
	if host.CustomDomain == "" {
		return ""
	} else {
		if host.CustomDomainMatcher == 1 && host.Node.AllowDomainMatcher == 1 {
			return fmt.Sprintf("*.%s", host.CustomDomain)
		} else {
			return host.CustomDomain
		}
	}
}
