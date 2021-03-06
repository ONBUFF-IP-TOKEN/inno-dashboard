package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	baseconf "github.com/ONBUFF-IP-TOKEN/baseapp/config"
)

var once sync.Once
var currentConfig *ServerConfig

type App struct {
	ApplicationName        string `json:"application_name" yaml:"application_name"`
	APIDocs                bool   `json:"api_docs" yaml:"api_docs"`
	CachePointExpiryPeriod int64  `json:"cache_point_expiry_period" yaml:"cache_point_expiry_period"`
	LiquidityUpdate        bool   `json:"liquidity_update" yaml:"liquidity_update"`
}

type Otp struct {
	EnableSwap bool   `json:"enable_swap" yaml:"enable_swap"`
	IssueName  string `json:"issue_name" yaml:"issue_name"`
}

type ApiAuth struct {
	AuthEnable    bool   `yaml:"auth_enable"`
	ApiAuthDomain string `json:"api_auth_domain" yaml:"api_auth_domain"`
	ApiAuthVerify string `json:"api_auth_verify" yaml:"api_auth_verify"`
}

type ApiPointManagerServer struct {
	InternalpiDomain string `yaml:"api_internal_domain"`
	ExternalpiDomain string `yaml:"api_external_domain"`
	InternalVer      string `yaml:"internal_ver"`
	ExternalVer      string `yaml:"external_ver"`
}

type ServerConfig struct {
	baseconf.Config `yaml:",inline"`

	App                App                   `yaml:"app"`
	Otp                Otp                   `yaml:"otp"`
	MssqlDBAccountAll  baseconf.DBAuth       `yaml:"mssql_db_account"`
	MssqlDBAccountRead baseconf.DBAuth       `yaml:"mssql_db_account_read"`
	MssqlDBLogRead     baseconf.DBAuth       `yaml:"mssql_db_log_read"`
	Auth               ApiAuth               `yaml:"api_auth"`
	PointMgrServer     ApiPointManagerServer `yaml:"api_point_manager_server"`
}

func GetInstance(filepath ...string) *ServerConfig {
	once.Do(func() {
		if len(filepath) <= 0 {
			panic(baseconf.ErrInitConfigFailed)
		}
		currentConfig = &ServerConfig{}
		if err := baseconf.Load(filepath[0], currentConfig); err != nil {
			currentConfig = nil
		} else {
			if os.Getenv("ASPNETCORE_PORT") != "" {
				port, _ := strconv.ParseInt(os.Getenv("ASPNETCORE_PORT"), 10, 32)
				currentConfig.APIServers[0].Port = int(port)
				currentConfig.APIServers[1].Port = int(port)
				fmt.Println(port)
			}
		}
	})

	return currentConfig
}
