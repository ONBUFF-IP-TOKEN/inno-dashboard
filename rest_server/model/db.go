package model

import (
	"strconv"

	"github.com/ONBUFF-IP-TOKEN/baseapp/base"
	baseconf "github.com/ONBUFF-IP-TOKEN/baseapp/config"
	"github.com/ONBUFF-IP-TOKEN/basedb"
	"github.com/ONBUFF-IP-TOKEN/baseutil/log"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/config"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/controllers/context"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/controllers/resultcode"
)

type PointDB struct {
	DatabaseID   int64
	DatabaseName string
	ServerName   string
}

type Point struct {
	PointIds []int64
}

type AppCoin struct {
	AppID int64 `json:"app_id"`
	context.CoinInfo
}

type AppInfo struct {
	AppId   int64  `json:"app_id,omitempty"`
	AppName string `json:"app_name,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type DB struct {
	MssqlAccountAll  *basedb.Mssql
	MssqlAccountRead *basedb.Mssql
	MssqlLogRead     *basedb.Mssql
	Cache            *basedb.Cache

	MssqlPoints map[int64]*basedb.Mssql

	ScanPointsMap map[int64]*context.PointInfo // 전체 포인트 종류 1 : key PointId
	ScanPoints    context.PointList            // 전체 포인트 종류 2

	AppPointsMap map[int64]*context.AppPointInfo // 전체 app과 포인트 1 : key appId
	AppPoints    context.AppPoints               // 전체 app과 포인트 2

	AppCoins map[int64][]*AppCoin // 전체 app에 속한 CoinID 정보 : key AppId

	CoinsMap map[int64]*context.CoinInfo // 전체 coin 정보 1 : key CoinId
	Coins    context.CoinList            // 전체 coin 정보 2
}

var gDB *DB

func GetDB() *DB {
	return gDB
}

func InitDB(conf *config.ServerConfig) (err error) {
	cache := basedb.GetCache(&conf.Cache)
	gDB = &DB{
		Cache: cache,
	}

	gDB.MssqlAccountAll, err = gDB.ConnectDB(&conf.MssqlDBAccountAll)
	if err != nil {
		return err
	}

	gDB.MssqlAccountRead, err = gDB.ConnectDB(&conf.MssqlDBAccountRead)
	if err != nil {
		return err
	}

	gDB.MssqlLogRead, err = gDB.ConnectDB(&conf.MssqlDBLogRead)
	if err != nil {
		return err
	}

	LoadDBPoint()
	return nil
}

func LoadDBPoint() {
	gDB.ScanPointsMap = make(map[int64]*context.PointInfo)
	gDB.AppCoins = make(map[int64][]*AppCoin)
	gDB.AppPointsMap = make(map[int64]*context.AppPointInfo)
	gDB.CoinsMap = make(map[int64]*context.CoinInfo)

	gDB.GetPointList()
	gDB.GetAppCoins()
	gDB.GetCoins()
	gDB.GetApps()
	gDB.GetAppPoints()
}

func MakeDbError(resp *base.BaseResponse, errCode int, err error) {
	resp.Return = errCode
	resp.Message = resultcode.ResultCodeText[errCode] + " : " + err.Error()
}

func (o *DB) ConnectDB(conf *baseconf.DBAuth) (*basedb.Mssql, error) {
	port, _ := strconv.ParseInt(conf.Port, 10, 32)
	mssqlDB, err := basedb.NewMssql(conf.Database, "", conf.ID, conf.Password, conf.Host, int(port))
	if err != nil {
		log.Errorf("err: %v, val: %v, %v, %v, %v, %v, %v",
			err, conf.Host, conf.ID, conf.Password, conf.Database, conf.PoolSize, conf.IdleSize)
		return nil, err
	}

	return mssqlDB, nil
}
