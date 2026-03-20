package model

import "time"

type Domain struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	CNAME     string    `json:"cname"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Origin struct {
	ID        int64     `json:"id"`
	DomainID  int64     `json:"domain_id"`
	Addr      string    `json:"addr"`
	Port      int       `json:"port"`
	Weight    int       `json:"weight"`
	Priority  int       `json:"priority"`
	Protocol  string    `json:"protocol"`
	CreatedAt time.Time `json:"created_at"`
}

type CacheRule struct {
	ID          int64     `json:"id"`
	DomainID    int64     `json:"domain_id"`
	PathPattern string    `json:"path_pattern"`
	TTL         int       `json:"ttl"`
	IgnoreQuery bool      `json:"ignore_query"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

type Certificate struct {
	ID        int64     `json:"id"`
	DomainID  int64     `json:"domain_id"`
	CertPEM   string    `json:"cert_pem,omitempty"`
	KeyPEM    string    `json:"-"`
	Issuer    string    `json:"issuer"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	AutoRenew bool      `json:"auto_renew"`
	CreatedAt time.Time `json:"created_at"`
}

type Node struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	IP            string     `json:"ip"`
	Region        string     `json:"region"`
	ISP           string     `json:"isp"`
	Status        string     `json:"status"`
	MaxBandwidth  int64      `json:"max_bandwidth"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at"`
}

type PurgeTask struct {
	ID          int64      `json:"id"`
	Type        string     `json:"type"`
	Targets     []string   `json:"targets"`
	Domain      string     `json:"domain"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/Response types

type CreateDomainReq struct {
	Domain string `json:"domain" binding:"required"`
	CNAME  string `json:"cname"`
}

type UpdateDomainReq struct {
	Status string `json:"status"`
	CNAME  string `json:"cname"`
}

type CreateOriginReq struct {
	Addr     string `json:"addr" binding:"required"`
	Port     int    `json:"port"`
	Weight   int    `json:"weight"`
	Priority int    `json:"priority"`
	Protocol string `json:"protocol"`
}

type UpdateOriginReq struct {
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	Weight   int    `json:"weight"`
	Priority int    `json:"priority"`
	Protocol string `json:"protocol"`
}

type CreateCacheRuleReq struct {
	PathPattern string `json:"path_pattern"`
	TTL         int    `json:"ttl" binding:"required"`
	IgnoreQuery bool   `json:"ignore_query"`
	Priority    int    `json:"priority"`
}

type PurgeReq struct {
	Targets []string `json:"targets" binding:"required"`
	Domain  string   `json:"domain" binding:"required"`
}

type PrefetchReq struct {
	URLs   []string `json:"urls" binding:"required"`
	Domain string   `json:"domain" binding:"required"`
}

type UploadCertReq struct {
	CertPEM   string `json:"cert_pem" binding:"required"`
	KeyPEM    string `json:"key_pem" binding:"required"`
	AutoRenew bool   `json:"auto_renew"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token string `json:"token"`
}
