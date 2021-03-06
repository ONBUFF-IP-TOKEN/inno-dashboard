package model

import (
	contextR "context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/ONBUFF-IP-TOKEN/baseutil/log"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/controllers/context"
	orginMssql "github.com/denisenkom/go-mssqldb"
)

const (
	USPAU_GetList_AccountCoins  = "[dbo].[USPAU_GetList_AccountCoins]"
	USPAU_GetList_AccountPoints = "[dbo].[USPAU_GetList_AccountPoints]"
	USPAU_GetList_Members       = "[dbo].[USPAU_GetList_Members]"
)

// 계정 코인 조회
func (o *DB) GetListAccountCoins(auid int64) ([]*context.MeCoin, error) {
	var returnValue orginMssql.ReturnStatus
	rows, err := o.MssqlAccountRead.QueryContext(contextR.Background(), USPAU_GetList_AccountCoins,
		sql.Named("AUID", auid),
		&returnValue)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		log.Error("USPAU_GetList_AccountCoins QueryContext err : ", err)
		return nil, err
	}

	meCoinList := []*context.MeCoin{}
	for rows.Next() {
		meCoin := &context.MeCoin{}
		if err := rows.Scan(&meCoin.CoinID,
			&meCoin.BaseCoinID,
			&meCoin.WalletAddress,
			&meCoin.Quantity,
			&meCoin.TodayAcqQuantity,
			&meCoin.TodayCnsmQuantity,
			&meCoin.TodayAcqExchangeQuantity,
			&meCoin.TodayCnsmExchangeQuantity,
			&meCoin.ResetDate); err != nil {
			log.Errorf("USPAU_GetList_AccountCoins Scan error %v", err)
			return nil, err
		} else {
			meCoin.CoinSymbol = o.CoinsMap[meCoin.CoinID].CoinSymbol
			meCoinList = append(meCoinList, meCoin)
		}
	}

	if returnValue != 1 {
		log.Errorf("USPAU_GetList_AccountCoins returnvalue error : %v", returnValue)
		return nil, errors.New("USPAU_GetList_AccountCoins returnvalue error " + strconv.Itoa(int(returnValue)))
	}
	return meCoinList, nil
}

// 계정 포인트 조회
func (o *DB) GetListAccountPoints(auid, muid int64) ([]*context.MePoint, error) {
	var returnValue orginMssql.ReturnStatus
	rows, err := o.MssqlAccountRead.QueryContext(contextR.Background(), USPAU_GetList_AccountPoints,
		sql.Named("AUID", auid),
		sql.Named("MUID", muid),
		&returnValue)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		log.Error("USPAU_GetList_AccountPoints QueryContext err : %v", err)
		return nil, err
	}

	var mePointList []*context.MePoint

	for rows.Next() {
		mePoint := context.MePoint{}
		if err := rows.Scan(&mePoint.AppID,
			&mePoint.PointID,
			&mePoint.TodayAcqQuantity,
			&mePoint.TodayCnsmQuantity,
			&mePoint.TodayAcqExchangeQuantity,
			&mePoint.TodayCnsmExchangeQuantity,
			&mePoint.ResetDate); err != nil {
			log.Errorf("USPAU_GetList_AccountPoints Scan error : %v", err)
			return nil, err
		} else {
			mePointList = append(mePointList, &mePoint)
		}
	}

	if returnValue != 1 {
		log.Errorf("USPAU_GetList_AccountPoints returnvalue error : %v", returnValue)
		return nil, errors.New("USPAU_GetList_AccountPoints returnvalue error " + strconv.Itoa(int(returnValue)))
	}
	return mePointList, nil
}

// 계정 앱 회원 조회
func (o *DB) GetListMembers(auid int64) ([]*context.Member, map[int64]*context.Member, error) {
	var returnValue orginMssql.ReturnStatus
	rows, err := o.MssqlAccountRead.QueryContext(contextR.Background(), USPAU_GetList_Members,
		sql.Named("AUID", auid),
		&returnValue)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		log.Error("USPAU_GetList_Members QueryContext err : %v", err)
		return nil, nil, err
	}

	var memberList []*context.Member
	memberMap := make(map[int64]*context.Member)

	for rows.Next() {
		member := context.Member{}
		if err := rows.Scan(&member.MUID, &member.AppID, &member.DatabaseID); err != nil {
			log.Errorf("USPAU_GetList_Members Scan error : %v", err)
			return nil, nil, err
		} else {
			memberMap[member.AppID] = &member
			memberList = append(memberList, &member)
		}
	}

	if returnValue != 1 {
		log.Errorf("USPAU_GetList_Members returnvalue error : %v", returnValue)
		return nil, nil, errors.New("USPAU_GetList_Members returnvalue error " + strconv.Itoa(int(returnValue)))
	}
	return memberList, memberMap, nil
}
