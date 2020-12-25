package game

import (
	"diceserver/game/dice"
	"diceserver/protocol"
	"github.com/lonng/nano/session"
	log "github.com/sirupsen/logrus"
)

type Loser struct {
	uid   int64
	score int
}

type Player struct {
	uid  int64  // 用户ID
	head string // 头像地址
	name string // 玩家名字
	ip   string // ip地址
	sex  int    // 性别
	coin int64  // 虚拟币

	// 玩家数据
	session *session.Session

	// 游戏相关字段
	onHand     dice.Dice
	calledDice string
	ctx        *dice.Context

	desk  *Desk // 当前桌
	turn  int   // 当前玩家在桌子的方位
	score int   // 玩家分数，默认为1000

	logger *log.Entry
}

func newPlayer(s *session.Session, uid int64, name, head, ip string, sex int) *Player {
	p := &Player{
		uid:   uid,
		name:  name,
		head:  head,
		ctx:   &dice.Context{Uid: uid},
		ip:    ip,
		sex:   sex,
		score: 1000,

		logger: log.WithField(fieldPlayer, uid),
	}

	p.ctx.Reset()
	p.bindSession(s)
	//p.syncCoinFromDB()

	return p
}

func (p *Player) setDesk(d *Desk, turn int) {
	if d == nil {
		p.logger.Error("桌号为空")
		return
	}

	p.desk = d
	p.turn = turn

	p.logger = log.WithFields(log.Fields{fieldDesk: p.desk.roomNo, fieldPlayer: p.uid})
}

func (p *Player) setIp(ip string) {
	p.ip = ip
}

func (p *Player) bindSession(s *session.Session) {
	p.session = s
	p.session.Set(kCurPlayer, p)
}

func (p *Player) removeSession() {
	p.session.Remove(kCurPlayer)
	p.session = nil
}

func (p *Player) Uid() int64 {
	return p.uid
}

func (p *Player) addDice(ids dice.Points) {
	p.onHand = dice.FromID(ids)
	p.logger.Debugf("游戏开局，手上骰子数量：%d 点数：%v", len(p.handPoints()), p.handPoints())
}

func (p *Player) callNum() string {
	var did string
	return did
}

func (p *Player) callDice() string {
	return p.calledDice
}

func (p *Player) call(calledDice string) {
	p.logger.Debugf("玩家选择: Dices=%s", calledDice)
	call := &protocol.Call{
		Uid:        p.Uid(),
		CalledDice: calledDice,
	}
	if err := p.desk.group.Broadcast("onCall", call); err != nil {
		log.Error(err)
	}
}

func (p *Player) showCheck(turn int, beturn int, lastCalledDice string) {

	ck := &protocol.ShowCheck{
		Uid:        p.Uid(),
		Open:       turn,
		Call:       beturn,
		CalledDice: lastCalledDice,
	}

	if err := p.desk.group.Broadcast("onCheck", ck); err != nil {
		log.Error(err)
	}
}

func (p *Player) showOpen(lastCalledDice string) {
	op := &protocol.ShowOpen{
		Uid:        p.Uid(),
		CalledDice: lastCalledDice,
	}
	println("curTurn:", p.desk.curTurn)
	if p.desk.curTurn != p.turn {
		if err := p.desk.group.Broadcast("onOpen", op); err != nil {
			log.Error(err)
		}
	}
}

func (p *Player) showWait(lastCalledDice string) {
	wo := &protocol.ShowWait{
		Uid:        p.Uid(),
		CalledDice: lastCalledDice,
	}
	if err := p.desk.group.Broadcast("onWait", wo); err != nil {
		log.Error(err)
	}
}

func (p *Player) handPoints() dice.Dice {
	return p.onHand
}

func (p *Player) calledPoint() string {
	return p.calledDice
}

func (p *Player) reset() {
	p.onHand = dice.Dice{}
	p.calledDice = ""

	//重置channel
	//close(p.chOperation)
	//p.chOperation = make(chan *protocol.OpChoosed, 1)
	p.ctx.Reset()
}

//断线重连后，同步牌桌数据
func (p *Player) syncDeskData() error {
	desk := p.desk
	data := &protocol.SyncDesk{
		Status:    desk.status(),
		Players:   []protocol.DeskPlayerData{},
		ScoreInfo: []protocol.ScoreInfo{},
	}

	markerUid := int64(0)
	lastCallUid := int64(0)
	for i, player := range desk.players {
		uid := player.Uid()
		if i == desk.bankerTurn {
			markerUid = uid
		}
		if i == desk.curTurn {
			lastCallUid = uid
		}

		score := protocol.ScoreInfo{
			Uid:   uid,
			Score: player.score,
		}

		data.ScoreInfo = append(data.ScoreInfo, score)
	}
	data.MarkerUid = markerUid
	data.LastCallUid = lastCallUid

	return p.session.Push("onSyncDesk", data)
}
