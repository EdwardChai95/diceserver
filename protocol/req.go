package protocol

import (
	"diceserver/pkg/constant"
)

type ReJoinDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReJoinDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type ReEnterDeskRequest struct {
	DeskNo string `json:"deskId"`
}

type ReEnterDeskResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}
type JoinDeskRequest struct {
	Version string `json:"version"`
	//AccountId int64         `json:"acId"`
	DeskNo string `json:"deskId"`
}

type TableInfo struct {
	DeskNo string              `json:"deskId"`
	Title  string              `json:"title"`
	Desc   string              `json:"desc"`
	Status constant.DeskStatus `json:"status"`
	Mode   int                 `json:"mode"`
}

type JoinDeskResponse struct {
	Code      int       `json:"code"`
	Error     string    `json:"error"`
	TableInfo TableInfo `json:"tableInfo"`
}

type DestoryDeskRequest struct {
	DeskNo string `json:"deskId"`
}

//选择执行的动作
type OpChoosed struct {
	Type       int
	CalledDice string
}

type MingAction struct {
	KouIndexs []int `json:"kou"` //index
	ChuPaiID  int   `json:"chu"`
	HuIndexs  []int `json:"hu"`
}

type OpChooseRequest struct {
	OpType int    `json:"optype"`
	Index  string `json:"idx"`
}

type OnOpenRequest struct {
	Turn       int    `json:"turn"`
	Beturn     int    `json:"beturn"`
	CalledDice string `json:"data"`
}

type OnCallRequest struct {
	CalledDice string `json:"data"`
}

type OnGuoRequest struct {
	CalledDice string `json:"data"`
}

type ChooseOneScoreRequest struct {
	Pos int `json:"pos"`
}

type DissolveStatusItem struct {
	DeskPos int    `json:"deskPos"`
	Status  string `json:"status"`
}

type DissolveResponse struct {
	DissolveUid    int64                `json:"dissolveUid"`
	DissolveStatus []DissolveStatusItem `json:"dissolveStatus"`
	RestTime       int32                `json:"restTime"`
}

type DissolveStatusRequest struct {
	Result bool `json:"result"`
}

type DissolveStatusResponse struct {
	DissolveStatus []DissolveStatusItem `json:"dissolveStatus"`
	RestTime       int32                `json:"restTime"`
}

type DissolveResult struct {
	DeskPos int `json:"deskPos"`
}

type PlayerOfflineStatus struct {
	Uid     int64 `json:"uid"`
	Offline bool  `json:"offline"`
}

type CoinChangeInformation struct {
	Coin int64 `json:"coin"`
}
