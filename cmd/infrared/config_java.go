package main

import (
	"net"
	"net/http"
	"time"

	"github.com/haveachin/infrared/internal/app/infrared"
	"github.com/haveachin/infrared/internal/pkg/java"
	"github.com/haveachin/infrared/pkg/webhook"
	"github.com/spf13/viper"
)

type JavaProxyConfig struct{}

func (cfg JavaProxyConfig) LoadGateways() ([]infrared.Gateway, error) {
	vpr := viper.Sub("defaults.java.gateway")

	var gateways []infrared.Gateway
	for id, v := range viper.GetStringMap("java.gateways") {
		vMap := v.(map[string]interface{})
		if err := vpr.MergeConfigMap(vMap); err != nil {
			return nil, err
		}
		var cfg javaGatewayConfig
		if err := vpr.Unmarshal(&cfg); err != nil {
			return nil, err
		}
		gateways = append(gateways, newJavaGateway(id, cfg))
	}

	return gateways, nil
}

func (cfg JavaProxyConfig) LoadServers() ([]infrared.Server, error) {
	vpr := viper.Sub("defaults.java.server")

	var servers []infrared.Server
	for id, v := range viper.GetStringMap("java.servers") {
		vMap := v.(map[string]interface{})
		if err := vpr.MergeConfigMap(vMap); err != nil {
			return nil, err
		}
		var cfg javaServerConfig
		if err := vpr.Unmarshal(&cfg); err != nil {
			return nil, err
		}
		servers = append(servers, newJavaServer(id, cfg))
	}

	return servers, nil
}

func (cfg JavaProxyConfig) LoadCPNs() ([]infrared.CPN, error) {
	var cpnCfg javaCpnConfig
	if err := viper.UnmarshalKey("java.processing_nodes", &cpnCfg); err != nil {
		return nil, err
	}

	cpns := make([]infrared.CPN, cpnCfg.Count)
	for n := range cpns {
		cpns[n].ConnProcessor = java.ConnProcessor{}
	}

	return cpns, nil
}

func (cfg JavaProxyConfig) LoadWebhooks() ([]webhook.Webhook, error) {
	vpr := viper.Sub("defaults.java.webhook")

	var webhooks []webhook.Webhook
	for id, v := range viper.GetStringMap("java.webhooks") {
		vMap := v.(map[string]interface{})
		if err := vpr.MergeConfigMap(vMap); err != nil {
			return nil, err
		}
		var cfg javaWebhookConfig
		if err := vpr.Unmarshal(&cfg); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, newJavaWebhook(id, cfg))
	}

	return webhooks, nil
}

type javaServerConfig struct {
	Domains           []string                  `mapstructure:"domains"`
	Address           string                    `mapstructure:"address"`
	ProxyBind         string                    `mapstructure:"proxy_bind"`
	DialTimeout       time.Duration             `mapstructure:"dial_timeout"`
	SendProxyProtocol bool                      `mapstructure:"send_proxy_protocol"`
	SendRealIP        bool                      `mapstructure:"send_real_ip"`
	DisconnectMessage string                    `mapstructure:"disconnect_message"`
	OnlineStatus      javaOnlineServerStatusConfig  `mapstructure:"online_status"`
	OfflineStatus     javaOfflineServerStatusConfig `mapstructure:"offline_status"`
}

type javaOnlineServerStatusConfig struct {
	VersionName    *string                          `mapstructure:"version_name,omitempty"`
	ProtocolNumber *int                             `mapstructure:"protocol_number,omitempty"`
	MaxPlayerCount *int                             `mapstructure:"max_player_count,omitempty"`
	PlayerCount    *int                             `mapstructure:"player_count,omitempty"`
	PlayerSample   []javaServerStatusPlayerSampleConfig `mapstructure:"player_sample,omitempty"`
	IconPath       *string                          `mapstructure:"icon_path,omitempty"`
	MOTD           *string                          `mapstructure:"motd,omitempty"`
}

type javaOfflineServerStatusConfig struct {
	VersionName    string                           `mapstructure:"version_name"`
	ProtocolNumber int                              `mapstructure:"protocol_number"`
	MaxPlayerCount int                              `mapstructure:"max_player_count"`
	PlayerCount    int                              `mapstructure:"player_count"`
	PlayerSample   []javaServerStatusPlayerSampleConfig `mapstructure:"player_sample"`
	IconPath       string                           `mapstructure:"icon_path"`
	MOTD           string                           `mapstructure:"motd"`
}

type javaServerStatusPlayerSampleConfig struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

type javaGatewayConfig struct {
	Binds                 []string      `mapstructure:"binds"`
	ReceiveProxyProtocol  bool          `mapstructure:"receive_proxy_protocol"`
	ReceiveRealIP         bool          `mapstructure:"receive_real_ip"`
	ClientTimeout         time.Duration `mapstructure:"client_timeout"`
	Servers               []string      `mapstructure:"servers"`
	ServerNotFoundMessage string        `mapstructure:"server_not_found_message"`
}

type javaCpnConfig struct {
	Count int `mapstructure:"count"`
}

type javaWebhookConfig struct {
	ClientTimeout time.Duration `mapstructure:"client_timeout"`
	URL           string        `mapstructure:"url"`
	Events        []string      `mapstructure:"events"`
}

func newJavaGateway(id string, cfg javaGatewayConfig) infrared.Gateway {
	return &java.Gateway{
		ID:                    id,
		Binds:                 cfg.Binds,
		ReceiveProxyProtocol:  cfg.ReceiveProxyProtocol,
		ReceiveRealIP:         cfg.ReceiveRealIP,
		ClientTimeout:         cfg.ClientTimeout,
		ServerIDs:             cfg.Servers,
		ServerNotFoundMessage: cfg.ServerNotFoundMessage,
	}
}

func newJavaServer(id string, cfg javaServerConfig) infrared.Server {
	return &java.Server{
		ID:      id,
		Domains: cfg.Domains,
		Dialer: net.Dialer{
			Timeout: cfg.DialTimeout,
			LocalAddr: &net.TCPAddr{
				IP: net.ParseIP(cfg.ProxyBind),
			},
		},
		Address:           cfg.Address,
		SendProxyProtocol: cfg.SendProxyProtocol,
		SendRealIP:        cfg.SendRealIP,
		DisconnectMessage: cfg.DisconnectMessage,
		OnlineStatus:      newJavaOnlineServerStatus(cfg.OnlineStatus),
		OfflineStatus:     newJavaOfflineServerStatus(cfg.OfflineStatus),
	}
}

func newJavaOnlineServerStatus(cfg javaOnlineServerStatusConfig) java.OnlineStatusResponse {
	return java.OnlineStatusResponse{
		VersionName:    cfg.VersionName,
		ProtocolNumber: cfg.ProtocolNumber,
		MaxPlayerCount: cfg.MaxPlayerCount,
		PlayerCount:    cfg.PlayerCount,
		IconPath:       cfg.IconPath,
		MOTD:           cfg.MOTD,
		PlayerSamples:  newJavaServerStatusPlayerSample(cfg.PlayerSample),
	}
}

func newJavaOfflineServerStatus(cfg javaOfflineServerStatusConfig) java.OfflineStatusResponse {
	return java.OfflineStatusResponse{
		VersionName:    cfg.VersionName,
		ProtocolNumber: cfg.ProtocolNumber,
		MaxPlayerCount: cfg.MaxPlayerCount,
		PlayerCount:    cfg.PlayerCount,
		IconPath:       cfg.IconPath,
		MOTD:           cfg.MOTD,
		PlayerSamples:  newJavaServerStatusPlayerSample(cfg.PlayerSample),
	}
}

func newJavaServerStatusPlayerSample(cfgs []javaServerStatusPlayerSampleConfig) []java.PlayerSample {
	playerSamples := make([]java.PlayerSample, len(cfgs))
	for n, cfg := range cfgs {
		playerSamples[n] = java.PlayerSample{
			Name: cfg.Name,
			UUID: cfg.UUID,
		}
	}
	return playerSamples
}

func newJavaWebhook(id string, cfg javaWebhookConfig) webhook.Webhook {
	return webhook.Webhook{
		ID: id,
		HTTPClient: &http.Client{
			Timeout: cfg.ClientTimeout,
		},
		URL:        cfg.URL,
		EventTypes: cfg.Events,
	}
}