package model

type GostClientForward struct {
	Base
	Name           string     `gorm:"column:name;index;comment:名称"`
	TargetIp       string     `gorm:"column:target_ip;index;comment:内网IP"`
	TargetPort     string     `gorm:"column:target_port;index;comment:内网端口"`
	Port           string     `gorm:"column:port;comment:访问端口"`
	ProxyProtocol  int        `gorm:"column:proxy_protocol;size:1;default:0;comment:代理协议"`
	NodeCode       string     `gorm:"column:node_code;index;comment:节点编号"`
	BackupNodeCode string     `gorm:"column:backup_node_code;index;comment:备用节点编号"`
	FailoverEnable int        `gorm:"column:failover_enable;size:1;default:2;comment:故障转移"`
	Node           GostNode   `gorm:"foreignKey:NodeCode;references:Code"`
	ClientCode     string     `gorm:"column:client_code;index;comment:客户端编号"`
	Client         GostClient `gorm:"foreignKey:ClientCode;references:Code"`
	UserCode       string     `gorm:"column:user_code;index;comment:用户编号"`
	User           SystemUser `gorm:"foreignKey:UserCode;references:Code"`
	Enable         int        `gorm:"column:enable;size:1;default:1;comment:启用状态"`
	Status         int        `gorm:"column:status;size:1;default:1;comment:状态"`
	BanReason      string     `gorm:"column:ban_reason;size:500;default:'';not null;comment:封禁原因"`
	MatcherEnable  int        `gorm:"column:matcher_enable;size:1;default:2;comment:是否开启匹配规则"`
	Matcher        string     `gorm:"column:matcher;comment:匹配规则"`
	TcpMatcher     string     `gorm:"column:tcp_matcher;comment:规则匹配"`
	SSHMatcher     string     `gorm:"column:ssh_matcher;comment:规则匹配"`
	UseEncryption  int        `gorm:"column:use_encryption;size:1;default:1;comment:加密"`
	UseCompression int        `gorm:"column:use_compression;size:1;default:1;comment:压缩"`
	PoolCount      int        `gorm:"column:pool_count;default:0;comment:复用连接数量"`
	Multiplexer         string `gorm:"column:multiplexer;comment:复用器httpconnect/tcpmux"`
	RouteByHTTPUser     string `gorm:"column:route_by_http_user;comment:按HTTP用户路由"`
	BandwidthLimit      string `gorm:"column:bandwidth_limit;comment:带宽限制如1MB(空=套餐限速)"`
	BandwidthMode       string `gorm:"column:bandwidth_mode;comment:带宽限制端client/server"`
	PortRange           string `gorm:"column:port_range;comment:范围端口如8000-8010"`
	LoadBalancerGroup   string `gorm:"column:load_balancer_group;comment:负载均衡组"`
	LoadBalancerGroupKey string `gorm:"column:load_balancer_group_key;comment:负载均衡组密钥"`
	HealthCheckType     string `gorm:"column:health_check_type;comment:健康检查类型tcp/http"`
	HealthCheckInterval int    `gorm:"column:health_check_interval;default:0;comment:健康检查间隔秒"`
	HealthCheckTimeout  int    `gorm:"column:health_check_timeout;default:0;comment:健康检查超时秒"`
	HealthCheckMaxFailed int   `gorm:"column:health_check_max_failed;default:0;comment:健康检查最大失败次数"`
	GostClientAdmission
	GostClientConfig
}
