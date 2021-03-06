package commonapi

import (
	"net/http"

	"github.com/ONBUFF-IP-TOKEN/baseapp/base"
	"github.com/ONBUFF-IP-TOKEN/baseutil/log"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/controllers/context"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/controllers/resultcode"
	"github.com/ONBUFF-IP-TOKEN/inno-dashboard/rest_server/model"
	"github.com/labstack/echo"
)

// 전체 포인트 리스트 조회
func GetPointList(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	resp.Value = model.GetDB().ScanPoints.Points

	return c.JSON(http.StatusOK, resp)
}

// 전체 포인트 리스트 DB reload
func PostReLoadPointList(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	if err := model.GetDB().GetPointList(); err == nil {
		resp.Value = model.GetDB().ScanPoints.Points
	} else {
		resp.SetReturn(resultcode.Result_DBError)
	}

	return c.JSON(http.StatusOK, resp)
}

// 전체 AppPoint 리스트 조회
func GetAppList(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	resp.Value = model.GetDB().AppPoints.Apps

	return c.JSON(http.StatusOK, resp)
}

// 전체 App 리스트 DB reload
func PostReLoadAppList(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	if err := model.GetDB().GetApps(); err == nil {
		if err = model.GetDB().GetAppPoints(); err == nil {
			resp.Value = model.GetDB().AppPoints.Apps
		} else {
			resp.SetReturn(resultcode.Result_DBError)
		}
	} else {
		resp.SetReturn(resultcode.Result_DBError)
	}

	return c.JSON(http.StatusOK, resp)
}

// 전체 코인 리스트 조회
func GetCoinList(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	resp.Value = model.GetDB().Coins.Coins

	return c.JSON(http.StatusOK, resp)
}

func PostReloadCoinList(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	if err := model.GetDB().GetAppCoins(); err == nil {
		if err = model.GetDB().GetCoins(); err == nil {
			resp.Value = model.GetDB().Coins.Coins
		} else {
			resp.SetReturn(resultcode.Result_DBError)
		}
	} else {
		resp.SetReturn(resultcode.Result_DBError)
	}

	return c.JSON(http.StatusOK, resp)
}

// App 포인트 별 전체 당일 누적/전환량 정보 조회
func GetAppPointAll(c echo.Context, reqAppPoint *context.ReqAppPointDaily) error {
	resp := new(base.BaseResponse)
	resp.Success()

	dailys := []*context.ResAppPointDaily{}

	// 포인트 종류 만큼 루프 돌려서 당일 정보만 추출해서 배열로 응답한다.
	for _, app := range model.GetDB().AppPoints.Apps {
		appDaily := &context.ResAppPointDaily{
			AppID: app.AppId,
		}

		for _, point := range app.Points {
			pointDaily := new(context.ResPointDaily)
			pointDaily.PointID = point.PointId
			key := model.MakeLogKeyOfPoint(appDaily.AppID, point.PointId, "day")
			if pointLiqs, err := model.GetDB().ZRevRangeLogOfPoint(key, 0, 0); err != nil {
				log.Errorf("ZRevRangeLogOfPoint error : %v", err)
				resp.SetReturn(resultcode.Result_Get_App_Point_Liquidity_Error)
			} else {
				for _, pointLiq := range pointLiqs {
					pointDaily.TodayAcqQuantity = pointLiq.AcqQuantity
					if pointLiq.CnsmExchangeQuantity < 0 { // 전환량은 음수이기 때문에 임시로 양수로 전환해준다.
						pointLiq.CnsmExchangeQuantity = -pointLiq.CnsmExchangeQuantity
					}
					pointDaily.TodayAcqExchangeQuantity = pointLiq.CnsmExchangeQuantity
				}
			}
			appDaily.ResPointDailys = append(appDaily.ResPointDailys, pointDaily)
		}
		dailys = append(dailys, appDaily)
	}

	resp.Value = dailys

	return c.JSON(http.StatusOK, resp)
}

// App 포인트 단위 별 당일 누적/전환량 정보 조회
func GetAppPoint(c echo.Context, reqAppPoint *context.ReqAppPointDaily) error {
	resp := new(base.BaseResponse)
	resp.Success()

	reqPointLiquidity := &context.ReqPointLiquidity{
		AppID:    reqAppPoint.AppID,
		PointID:  reqAppPoint.PointID,
		Interval: 0,
	}

	if pointLiquiditys, err := model.GetDB().GetListDailyPoints(reqPointLiquidity); err != nil {
		resp.SetReturn(resultcode.Result_Get_App_Point_DailyLiquidity_Error)
	} else {
		if len(pointLiquiditys) == 0 { //오늘 누적된 포인터가 없을때 강제로 0으로 보정
			pointLiquiditys = append(pointLiquiditys, &context.PointLiquidity{
				AcqQuantity:          0,
				CnsmExchangeQuantity: 0,
			})
		}
		if pointLiquiditys[0].CnsmExchangeQuantity < 0 { // 전환량은 음수이기 때문에 임시로 양수로 전환해준다.
			pointLiquiditys[0].CnsmExchangeQuantity = -pointLiquiditys[0].CnsmExchangeQuantity
		}
		pointDaily := &context.ResPointDaily{
			PointID:                  reqPointLiquidity.PointID,
			TodayAcqQuantity:         pointLiquiditys[0].AcqQuantity,
			TodayAcqExchangeQuantity: pointLiquiditys[0].CnsmExchangeQuantity,
		}

		res := context.ResAppPointDaily{
			AppID: reqPointLiquidity.AppID,
		}
		res.ResPointDailys = append(res.ResPointDailys, pointDaily)
		resp.Value = res
	}

	return c.JSON(http.StatusOK, resp)
}

// 코인 별 당일 누적/전환량 조회
func GetAppCoinDailyAll(c echo.Context, reqAppCoinDaily *context.ReqAppCoinDaily) error {
	resp := new(base.BaseResponse)
	resp.Success()

	dailys := []*context.ResAppCoinDaily{}
	// 코인 종류 만큼 루프 돌려서 당일 정보만 추출해서 배열로 응답한다.
	for _, coin := range model.GetDB().Coins.Coins {
		key := model.MakeLogKeyOfCoin(coin.CoinId, "day")
		if coinLiqs, err := model.GetDB().ZRevRangeLogOfCoin(key, 0, 0); err != nil {
			log.Errorf("ZRevRangeLogOfCoin error : %v", err)
			resp.SetReturn(resultcode.Result_Get_App_Coin_Liquidity_Error)
		} else {
			if len(coinLiqs) == 0 {
				continue
			}
			if coinLiqs[0].CnsmExchangeQuantity < 0 {
				coinLiqs[0].CnsmExchangeQuantity = -coinLiqs[0].CnsmExchangeQuantity
			}

			coinDaily := &context.ResAppCoinDaily{
				CoinID:                   coin.CoinId,
				TodayAcqQuantity:         coinLiqs[0].AcqExchangeQuantity,
				TodayAcqExchangeQuantity: coinLiqs[0].CnsmExchangeQuantity,
			}
			dailys = append(dailys, coinDaily)
		}
	}

	resp.Value = dailys

	return c.JSON(http.StatusOK, resp)
}

// 코인 별 당일 누적/전환량 조회
func GetAppCoinDaily(c echo.Context, reqAppCoinDaily *context.ReqAppCoinDaily) error {
	resp := new(base.BaseResponse)
	resp.Success()

	reqCoinLiquidity := &context.ReqCoinLiquidity{
		CoinID:   reqAppCoinDaily.CoinID,
		Interval: 0,
	}

	if coinLiquiditys, err := model.GetDB().GetListDailyCoins(reqCoinLiquidity); err != nil {
		resp.SetReturn(resultcode.Result_Get_App_Coin_DailyLiquidity_Error)
	} else {
		if len(coinLiquiditys) == 0 { // 오늘 누적된 코인량이 없을때
			coinLiquiditys = append(coinLiquiditys, &context.CoinLiquidity{
				AcqExchangeQuantity:  0,
				CnsmExchangeQuantity: 0,
			})
		}
		if coinLiquiditys[0].CnsmExchangeQuantity < 0 {
			coinLiquiditys[0].CnsmExchangeQuantity = -coinLiquiditys[0].CnsmExchangeQuantity
		}
		res := context.ResAppCoinDaily{
			CoinID:                   reqAppCoinDaily.CoinID,
			TodayAcqQuantity:         coinLiquiditys[0].AcqExchangeQuantity,
			TodayAcqExchangeQuantity: coinLiquiditys[0].CnsmExchangeQuantity,
		}
		resp.Value = res
	}

	return c.JSON(http.StatusOK, resp)
}

// 하루 최대 수집 포인트 양 조회
func GetAppPointMax(c echo.Context) error {
	resp := new(base.BaseResponse)
	resp.Success()

	return c.JSON(http.StatusOK, resp)
}
