package protocol

import (
	"diceserver/pkg/constant"
)

type EnterDeskInfo struct {
	DeskPos  int    `json:"deskPos"`
	Uid      int64  `json:"acId"`
	Nickname string `json:"nickname"`
	IsReady  bool   `json:"isReady"`
	Sex      int    `json:"sex"`
	IsExit   bool   `json:"isExit"`
	HeadUrl  string `json:"headURL"`
	Score    int    `json:"score"`
	IP       string `json:"ip"`
	Offline  bool   `json:"offline"`
}

type ExitResponse struct {
	AccountId int64 `json:"acid"`
	IsExit    bool  `json:"isexit"`
	ExitType  int   `json:"exitType"`
	DeskPos   int   `json:"deskPos"`
}

type PlayerEnterDesk struct {
	Data []EnterDeskInfo `json:"data"`
}

type ExitRequest struct {
	IsDestroy bool `json:"isDestroy"`
}

type DeskBasicInfo struct {
	DeskID string `json:"deskId"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
	Mode   int    `json:"mode"`
}

type ScoreInfo struct {
	Uid   int64 `json:"acId"`
	Score int   `json:"score"`
}

type RollInfo struct {
	Uid    int64 `json:"acId"`
	OnHand []int `json:"dcs"`
}

type Roll struct {
	AccountInfo []RollInfo `json:"accountInfo"`
}

type Call struct {
	Uid        int64  `json:"acId"`
	CalledDice string `json:"dcs"`
}

type ShowCheck struct {
	Uid        int64      `json:"acId"`
	Open       int        `json:"open"`
	Call       int        `json:"call"`
	CalledDice string     `json:"dcs"`
}

type ShowOpen struct {
	Uid        int64  `json:"acId"`
	CalledDice string `json:"dcs"`
}

type ShowWait struct {
	Uid        int64  `json:"acId"`
	CalledDice string `json:"dcs"`
}

type MatchStats struct {
	TotalScore  int    `json:"totalScore"`  //总分
	Uid         int64  `json:"uid"`         //id
	Account     string `json:"account"`     //名字
	IsPaoWang   bool   `json:"isPaoWang"`   //是否是炮王
	IsBigWinner bool   `json:"isBigWinner"` //是否是大赢家
}

type OpenInfo struct {
	Uid          int64        `json:"acId"`
	OpenDiceType OpenDiceType `json:"OpenDiceType"`
	WinLose      bool         `json:"winlose"`
}

type HandDicesInfo struct {
	Uid   int64 `json:"acId"`
	Dices []int `json:"shouPai"`
}

type GameEndScoreChange struct {
	Uid    int64 `json:"acId"`
	Score  int   `json:"score"`
	Remain int   `json:"remain"`
}

type RoundOverStats struct {
	Title       string               `json:"title"`
	HandDices   []*HandDicesInfo     `json:"tiles"`
	ScoreChange []GameEndScoreChange `json:"scoreChange"`
}

type DeskPlayerData struct {
	Uid       int64 `json:"acId"`
	HandDices []int `json:"shouPaiIds"`
	CallDices []int `json:"callDices"`
	Open      int   `json:"open"`
	Score     int   `json:"score"`
}

type SyncDesk struct {
	Status      constant.DeskStatus `json:"status"` //1,2,3,4,5
	Players     []DeskPlayerData    `json:"players"`
	ScoreInfo   []ScoreInfo         `json:"scoreInfo"`
	MarkerUid   int64               `json:"markerAcId"`
	LastCallUid int64               `json:"lastCallAcId"`
}

type DeskOptions struct {
	Mode int `json:"mode"`
}

type CreateDeskRequest struct {
	Version  string       `json:"version"` //客户端版本
	ClubId   int64        `json:"clubId"`  // 俱乐部ID
	DeskOpts *DeskOptions `json:"options"` // 游戏额外选项
}

type CreateDeskResponse struct {
	Code      int       `json:"code"`
	Error     string    `json:"error"`
	TableInfo TableInfo `json:"tableInfo"`
}

type ReConnect struct {
	Uid     int64  `json:"uid"`
	Name    string `json:"name"`
	HeadUrl string `json:"headUrl"`
	Sex     int    `json:"sex"`
}

type DeskListRequest struct {
	Player int64 `json:"player"`
	Offset int   `json:"offset"`
	Count  int   `json:"count"`
}

type Desk struct {
	Id           int64  `json:"id"`
	DeskNo       string `json:"desk_no"`
	Mode         int    `json:"mode"`
	Player0      int64  `json:"player0"`
	Player1      int64  `json:"player1"`
	Player2      int64  `json:"player2"`
	Player3      int64  `json:"player3"`
	Player4      int64  `json:"player4"`
	PlayerName0  string `json:"player_name0"`
	PlayerName1  string `json:"player_name1"`
	PlayerName2  string `json:"player_name2"`
	PlayerName3  string `json:"player_name3"`
	PlayerName4  string `json:"player_name4"`
	ScoreChange0 int    `json:"score_change0"`
	ScoreChange1 int    `json:"score_change1"`
	ScoreChange2 int    `json:"score_change2"`
	ScoreChange3 int    `json:"score_change3"`
	ScoreChange4 int    `json:"score_change4"`
	CreatedAt    int64  `json:"created_at"`
	CreatedAtStr string `json:"created_at_str"`
	Extras       string `json:"extras"`
}

type DestroyDeskResponse struct {
	RoundStats       *RoundOverStats `json:"roundStats"`
	MatchStats       []MatchStats    `json:"stats"`
	Title            string          `json:"title"`
	IsNormalFinished bool            `json:"isNormalFinished"`
}

type DeskListResponse struct {
	Code  int    `json:"code"`
	Total int64  `json:"total"` //总数量
	Data  []Desk `json:"data"`
}

type DeleteDeskByIDRequest struct {
	ID string `json:"id"` //房间ID
}
type DeskByIDRequest struct {
	ID int64 `json:"id"` //房间ID
}

type DeskByIDResponse struct {
	Code int   `json:"code"`
	Data *Desk `json:"data"`
}

type RoundReady struct {
	Multiple int `json:"multiple"`
}

type UnCompleteDeskResponse struct {
	Exist     bool      `json:"exist"`
	TableInfo TableInfo `json:"tableInfo"`
}

type RecordingVoice struct {
	FileId string `json:"fileId"`
}

type PlayRecordingVoice struct {
	Uid    int64  `json:"uid"`
	FileId string `json:"fileId"`
}

type ClientInitCompletedRequest struct {
	IsReEnter bool `json:"isReenter"`
}

type EnterPublicDeskRequest struct {
	Version string `json:"version"`
	DeskNo  string `json:"deskId"`
	Mode    int    `json:"mode"`
}

type EnterPublicDeskResponse struct {
	Code      int       `json:"code"`
	Error     string    `json:"error"`
	TableInfo TableInfo `json:"tableInfo"`
}

type ExitPublicDeskRequest struct {
	DeskNo string `json:"deskId"`
}
