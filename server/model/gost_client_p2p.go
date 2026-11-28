package model

type GostClientP2P struct {
	Base
	Name           string     `gorm:"column:name;index;comment:名称"`
	TargetIp       string     `gorm:"column:target_ip;index;comment:内网IP"`
	TargetPort     string     `gorm:"column:target_port;index;comment:内网端口"`
	VKey           string     `gorm:"column:v_key;comment:访问密钥"`
	Forward        int        `gorm:"column:forward;size:1;default:0;comment:是否中继"`
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
	UseEncryption  int        `gorm:"column:use_encryption;size:1;default:1;comment:加密"`
	UseCompression int        `gorm:"column:use_compression;size:1;default:1;comment:压缩"`
	PoolCount      int        `gorm:"column:pool_count;default:0;comment:复用连接数量"`
	NatType        string     `gorm:"column:nat_type;comment:最近探测的NAT类型"`
	RelayBytes     int64      `gorm:"column:relay_bytes;default:0;comment:STCP中继累计流量字节"`
	KeepTunnelOpen int        `gorm:"column:keep_tunnel_open;size:1;default:2;comment:打洞保持隧道开启"`
	GostClientConfig
}
