package constant

type Behavior int

const (
	BehaviorNone Behavior = iota
	BehaviorPeng
	BehaviorGang
	BehaviorAnGang
	BehaviorBaGang
	BehaviorHu
)

type DeskStatus int32

const (
	//创建桌子
	DeskStatusStart DeskStatus = iota
	//发牌
	DeskStatusZhunBei
	//齐骰
	DeskStatusQiTou
	//游戏
	DeskStatusPlaying
	DeskStatusOver
	//游戏终/中止
	DeskStatusInterruption
	//已销毁
	DeskStatusDestory
	//已经清洗,即为下一轮准备好
	DeskStatusCleaned
)

var stringify = [...]string{
	DeskStatusStart:        "开始",
	DeskStatusZhunBei:      "准备",
	DeskStatusQiTou:		"齐骰",
	DeskStatusPlaying:      "游戏中",
	DeskStatusOver:         "单局完成",
	DeskStatusInterruption: "游戏终/中止",
	DeskStatusDestory:      "已销毁",
	DeskStatusCleaned:      "已清洗",
}

func (s DeskStatus) String() string {
	return stringify[s]
}
