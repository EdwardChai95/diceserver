package game

import (
	"diceserver/db"
	"diceserver/db/model"
	"diceserver/game/dice"
	"diceserver/game/history"
	"diceserver/pkg/constant"
	"diceserver/pkg/errutil"
	"diceserver/pkg/room"
	"diceserver/protocol"
	"fmt"
	"github.com/lonng/nano/scheduler"
	"strings"
	"sync/atomic"

	"github.com/lonng/nano"
	"github.com/lonng/nano/session"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	illegalTurn = -1
	deskOver    = "" // 桌子结束
)

type Desk struct {
	roomNo  room.Number         // 房间号
	deskID  int64               // desk表的pk
	mode    int                 // 房间人数
	state   constant.DeskStatus // 房间状态
	round   uint32              // 第n局
	players []*Player
	group   *nano.Group // 组播通道

	allDices   dice.Dice // 所有骰子
	bankerTurn int       // 庄家方位
	curTurn    int       // 当前方位
	isNewRound bool      // 是否是每局的第一次叫牌

	snapshot   *history.History   // 为本局准备快照
	matchStats history.MatchStats // 单局分值变化统计

	openPlayers map[int64]bool // 是否开牌
	openPlayer  int64          // 开牌的玩家

	dissolve *dissolveContext // 退出相关状态
	prepare  *prepareContext  // 准备相关状态

	lastCallDice string // 最后一次叫点数
	lastCallUid  int64  // 最后一个叫的玩家

	latestEnter *protocol.PlayerEnterDesk // 最新的进入状态

	logger *log.Entry
}

func NewDesk(roomNo room.Number, mode int) *Desk {
	d := &Desk{
		roomNo:     roomNo,
		mode:       mode,
		players:    []*Player{},
		group:      nano.NewGroup(uuid.New()),
		bankerTurn: turnUnknown,
		isNewRound: true,
		matchStats: make(history.MatchStats),
		prepare:    newPrepareContext(),
		logger:     log.WithField(fieldDesk, roomNo),
	}
	d.dissolve = newDissolveContext(d)
	return d
}

func (d *Desk) totalPlayerCount() int {
	return d.mode
}

func (d *Desk) totalDiceCount() int {
	if d.mode == 5 {
		return 25
	} else if d.mode == 4 {
		return 20
	} else if d.mode == 3 {
		return 15
	} else {
		return 10
	}
}

func (d *Desk) save() error {
	var name3, name4 string
	var player3, player4 int64
	if d.mode == ModeFours {
		name3 = d.players[3].name
		player3 = d.players[3].Uid()
	} else if d.mode == ModeFives {
		name3 = d.players[3].name
		player3 = d.players[3].Uid()
		name4 = d.players[4].name
		player4 = d.players[4].Uid()
	}
	// save to database
	desk := &model.Desk{
		Mode:        d.mode,
		DeskNo:      string(d.roomNo),
		Player0:     d.players[0].Uid(),
		Player1:     d.players[1].Uid(),
		Player2:     d.players[2].Uid(),
		Player3:     player3,
		Player4:     player4,
		PlayerName0: d.players[0].name,
		PlayerName1: d.players[1].name,
		PlayerName2: d.players[2].name,
		PlayerName3: name3,
		PlayerName4: name4,
	}
	d.logger.Infof("保存房间数据， 创建时间： %d", desk.CreatedAt)

	if err := db.UpdateNewDesk(desk); err != nil {
		return err
	}

	d.deskID = desk.Id
	return nil
}

// 如果是重新进入 isRejoin: true
func (d *Desk) playerJoin(s *session.Session, isRejoin bool) error {
	uid := s.UID()
	var (
		p   *Player
		err error
	)
	println("Uid:", uid)
	if isRejoin {
		d.dissolve.updateOnlineStatus(uid, true)
		p, err = d.playerWithId(uid)
		if err != nil {
			d.logger.Errorf("玩家：%d重新加入房间，但是没有找到玩家在房间内的数据", uid)
			return err
		}
		//加入分组
		d.group.Add(s)
	} else {
		exists := false
		for _, p := range d.players {

			if p.Uid() == uid {
				exists = true
				p.logger.Warn("玩家已经在房间中")
				break
			}
		}
		if !exists {
			p = s.Value(kCurPlayer).(*Player)
			d.players = append(d.players, p)
			for i, p := range d.players {
				p.setDesk(d, i)
			}
			//d.roundStats[uid] = &history.Record{}
		}
	}
	return nil
}

func (d *Desk) syncDeskStatus() {
	d.latestEnter = &protocol.PlayerEnterDesk{Data: []protocol.EnterDeskInfo{}}
	for i, p := range d.players {
		uid := p.Uid()
		d.latestEnter.Data = append(d.latestEnter.Data, protocol.EnterDeskInfo{
			DeskPos:  i,
			Uid:      uid,
			Nickname: p.name,
			IsReady:  d.prepare.isReady(uid),
			Sex:      p.sex,
			IsExit:   false,
			HeadUrl:  p.head,
			Score:    p.score,
			IP:       p.ip,
			Offline:  !d.dissolve.isOnline(uid),
		})
	}
	d.group.Broadcast("onPlayerEnter", d.latestEnter)
}

func (d *Desk) checkStart() {
	s := d.status()
	if s != constant.DeskStatusStart {
		d.logger.Infof("当前房间还有人未准备就绪，不能开始游戏")
		return
	}

	if count, num := len(d.players), d.totalPlayerCount(); count < num {
		d.logger.Infof("当前房间玩家数量不足，不能开始游戏，当前玩家=%d，最低人数=%d", count, num)
		return
	}
	for _, p := range d.players {
		if uid := p.Uid(); !d.prepare.isReady(uid) {
			p.logger.Info("玩家未准备")
			return
		}
	}
	d.start()
}

func (d *Desk) title() string {
	return strings.TrimSpace(fmt.Sprintf("房号: %s 模式: %d", d.roomNo, d.mode))
}

func (d *Desk) desc(detail bool) string {
	desc := []string{}
	mode := d.mode

	desc = append(desc, fmt.Sprintf("%d人场次", mode))

	return strings.Join(desc, "")
}

func (d *Desk) start() { //未添加
	d.setStatus(constant.DeskStatusZhunBei)

	var (
		totalPlayerCount = d.totalPlayerCount() // 玩家数量
		totalDiceCount   = d.totalDiceCount()   // 骰子数量
	)

	if err := d.save(); err != nil {
		d.logger.Error(err)
	}
	d.curTurn = 0
	// 桌面基本信息
	basic := &protocol.DeskBasicInfo{
		DeskID: d.roomNo.String(),
		Title:  d.title(),
		Desc:   d.desc(true),
		Mode:   d.mode,
	}

	d.group.Broadcast("onDeskBasicInfo", basic)
	allDices := dice.New(totalDiceCount)
	d.logger.Debugf("骰子数量=%d, 玩家数量=%d, 所有骰子=%v", totalDiceCount, totalPlayerCount, allDices)

	info := make([]protocol.RollInfo, totalPlayerCount)
	for i, p := range d.players {
		info[i] = protocol.RollInfo{
			Uid:    p.Uid(),
			OnHand: make([]int, 5),
		}
		nextIndex := (i + 1) * 5
		copy(info[i].OnHand, allDices[i*5:nextIndex])
	}

	d.allDices = make(dice.Dice, len(allDices))
	for i, id := range allDices {
		dice := dice.PointFromID(id)
		d.allDices[i] = dice
	}

	d.logger.Debugf("游戏开局， 骰子数量=%d, 所有骰子:%v", len(d.allDices), d.allDices)
	for turn, player := range d.players {
		player.addDice(info[turn].OnHand)
	}

	roll := &protocol.Roll{AccountInfo: info}

	d.group.Broadcast("onRoll", roll)

	name5 := "/"
	name4 := "/"
	if len(d.players) > 4 {
		name5 = d.players[4].name
		name4 = d.players[3].name
	} else if len(d.players) > 3 {
		name4 = d.players[3].name
	}
	d.snapshot = history.New(
		d.deskID,
		d.mode,
		d.players[0].name,
		d.players[1].name,
		d.players[2].name,
		name4,
		name5,
		basic,
		d.latestEnter,
		roll,
	)
}

func (d *Desk) rollDiceFinished(uid int64) error {
	if d.status() > constant.DeskStatusQiTou {
		d.logger.Debugf("当前牌桌状态: %s", d.status().String())
		return errutil.ErrIllegalDeskStatus
	}

	d.prepare.sorted(uid)

	// 等待所有人齐骰子
	for _, p := range d.players {
		if !d.prepare.isSorted(p.Uid()) {
			return nil
		}
	}

	d.setStatus(constant.DeskStatusQiTou)

	go d.play()

	return nil
}

func (d *Desk) nextTurn() {
	d.curTurn++
	d.curTurn = d.curTurn % d.totalPlayerCount()
}

/*
func (d *Desk) isRoundOver() bool {
	// 中断表示本局结束
	s := d.status()
	if s == constant.DeskStatusOver {
		return true
	}

	return len(d.openPlayer) == d.totalPlayerCount()-1
}
*/

// 循环中的核心逻辑
// 1. 叫牌
// 2. 检查是否有玩家开牌
func (d *Desk) play() {
	defer func() {
		if err := recover(); err != nil {
			d.logger.Errorf("Error=%v", err)
			println(stack())
		}
	}()

	d.setStatus(constant.DeskStatusPlaying)
	d.logger.Debug("开始游戏")

	/*curPlayer := d.players[d.curTurn] // 当前叫牌玩家，初始为庄家

	MAIN_LOOP:
		for !d.isRoundOver() {
			// 切换到下一个玩家
			if !d.isNewRound {
				d.nextTurn()
				curPlayer = d.currentPlayer()
			}

			//curPlayer.showCall()
			did := curPlayer.callNum()
			if did == deskOver {
				break MAIN_LOOP
			}

			d.lastCallDice = did
			d.lastCallUid = curPlayer.Uid()
			curPlayer.calledDice = did
			curPlayer.ctx.LastCallDice = did

			typ := d.showOpen(did)
			if typ == nil {
				break MAIN_LOOP
			}
		}*/

	if d.status() != constant.DeskStatusInterruption {
		d.setStatus(constant.DeskStatusOver)
	}

}

func (d *Desk) currentPlayer() *Player {
	return d.players[d.curTurn]
}

func (d *Desk) showOpen(calledDice string) error {
	// 叫牌玩家
	callPlayer := d.currentPlayer()
	curDice := calledDice
	playerCount := d.totalPlayerCount()
	turn := d.curTurn
	for {
		turn++
		turn = turn % playerCount
		otherPlayer := d.players[turn]
		if turn == d.curTurn {
			otherPlayer.showWait(curDice)
			d.nextTurn()
			break
		}

		if turn != callPlayer.turn {
			otherPlayer.showOpen(curDice)
		}
	}
	return nil
}

func (d *Desk) showCall(calledDice string) error {
	curDice := calledDice
	playerCount := d.totalPlayerCount()
	turn := d.curTurn
	for {
		turn++
		turn = turn % playerCount
		callPlayer := d.players[turn]

		if turn == d.curTurn {
			callPlayer.call(curDice)
			break
		}
	}

	return nil
}

func (d *Desk) showCheck(turn int, beturn int, calledDice string) error {
	callPlayer := d.currentPlayer()
	curDice := calledDice
	playerCount := d.totalPlayerCount()
	curturn := d.curTurn


	for {
		curturn++
		curturn = curturn % playerCount
		player := d.players[curturn]
		if curturn != callPlayer.turn {
			player.showCheck(turn, beturn, curDice)
		}

		if curturn == d.curTurn {
			player.showCheck(turn, beturn, curDice)
			break
		}

	}
	return nil
}

/*

func (d *Desk) roundOverStatsForPlayer(p *Player) *protocol.RoundStats {

}

func (d *Desk) roundOverHelper() *protocol.RoundOverStats {
	// 游戏结束
	overStats := &protocol.RoundOverStats{
		Title:       d.desc(false),
		HandDices:   []*protocol.HandDicesInfo{},
		ScoreChange: []protocol.GameEndScoreChange{},
	}

	if d.status() == constant.DeskStatusCleaned {
		return overStats
	}

	// 总结算分数
	for i, p := range d.players {
		uid := p.Uid()
		stats := d.roundOverStatsForPlayer(p)
		total := stats.Total
		p.score += total


	}

	return overStats
}*/

func (d *Desk) setStatus(s constant.DeskStatus) {
	atomic.StoreInt32((*int32)(&d.state), int32(s))
}

func (d *Desk) status() constant.DeskStatus {
	return constant.DeskStatus(atomic.LoadInt32((*int32)(&d.state)))
}

/*
func (d *Desk) roundOver() {
	stats := d.roundOverHelper()
	status := d.status()

	// 满场
	isMaxRound := d.round >= uint32(d.opts.MaxRound) && status == constant.DeskStatusOver

	// game over
	if status == constant.DeskStatusOver && !isMaxRound {
		d.group.Broadcast("onRoundEnd", stats)
		d.clean()
	} else {
		// 最后一句以及中断统计的GameEnd与场结算一起发送
		d.finalSettlement(isMaxRound, stats)
	}
}

func (d *Desk) clean() {
	d.state = constant.DeskStatusCleaned
	d.isNewRound = true

	d.openPlayers = map[int64]bool{}

	d.prepare.reset()

	//重置玩家状态
	for _, p := range d.players {
		//d.roundStats[p.Uid()] = &history.Record{}  // 存储历史
		p.reset()
	}
}

func (d *Desk) finalSettlement(isNormalFinished bool, ge *protocol.RoundOverStats) {
	d.logger.Debugf("本场游戏结束，结算数据: %#v", ge)

}*/

func (d *Desk) isRoundOver() bool {
	return d.status() == constant.DeskStatusOver
}

func (d *Desk) destroy() {
	for i := range d.players {
		p := d.players[i]
		d.logger.Debugf("销毁房间，清除玩家%d数据", p.Uid())
		p.reset()
		p.desk = nil
		p.turn = 0
		d.players[i] = nil
	}
	d.group.Close()
	d.prepare.reset()

	scheduler.PushTask(func() {
		defaultDeskManager.setDesk(d.roomNo, nil)
	})
}

/*
func (d *Desk) scoreChangeHelper(winner int64, losers []Loser, typ ScoreChangeType, callDice string) {
	// 向赢/输者队列添加信息
	d.scoreChangeForUid(winner, &scoreChangeInfo{
		score:    loser.score, // 赢了多少
		uid:      loser,       // 谁输的
		typ:      typ,
		callDice: callDice,
	})

	d.scoreChangeForUid(winner, &scoreChangeInfo{
		score:    -loser.score, // 输了多少
		uid:      winner,       // 谁赢的
		typ:      typ,
		callDice: callDice,
	})
}
*/
func (d *Desk) onPlayerExit(s *session.Session, isDisconnect bool) {
	uid := s.UID()
	d.group.Leave(s)
	if isDisconnect {
		//d.disslove.updateOnlineStatus(uid, false)
	} else {
		restPlayers := []*Player{}
		for _, p := range d.players {
			if p.Uid() != uid {
				restPlayers = append(restPlayers, p)
			} else {
				p.reset()
				p.desk = nil
				p.score = 1000
				p.turn = 0
			}
		}
		d.players = restPlayers
	}
}

func (d *Desk) playerWithId(uid int64) (*Player, error) {
	for _, p := range d.players {
		if p.Uid() == uid {
			return p, nil
		}
	}
	return nil, errutil.ErrPlayerNotFound
}

/*
func (d *Desk) setNextRoundBanker(uid int64, override bool) {
	// 如果已经设置了庄家，如果一炮双响则重新设置庄家
	if d.isMakerSet && !override {
		return
	}
	for i, p := range d.players {
		if p.Uid() == uid {
			d.bankerTurn = i
			break
		}
	}
	d.isMakerSet = true
}
*/
func (d *Desk) onPlayerReJoin(s *session.Session) error {
	// 同步房间基本信息
	basic := &protocol.DeskBasicInfo{
		DeskID: d.roomNo.String(),
		Title:  d.title(),
		Desc:   d.desc(true),
	}
	if err := s.Push("onDeskBasicInfo", basic); err != nil {
		log.Error(err.Error())
		return err
	}

	// 同步所有玩家数据
	enter := &protocol.PlayerEnterDesk{Data: []protocol.EnterDeskInfo{}}
	for i, p := range d.players {
		uid := p.Uid()
		enter.Data = append(enter.Data, protocol.EnterDeskInfo{
			DeskPos:  i,
			Uid:      uid,
			Nickname: p.name,
			IsReady:  d.prepare.isReady(uid),
			Sex:      p.sex,
			IsExit:   false,
			HeadUrl:  p.head,
			Score:    p.score,
			IP:       p.ip,
			Offline:  !d.dissolve.isOnline(uid),
		})
	}
	if err := s.Push("onPlayerEnter", enter); err != nil {
		log.Error(err.Error())
		return err
	}

	p, err := playerWithSession(s)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	if err := d.playerJoin(s, true); err != nil {
		log.Error(err)
	}

	// 游戏结束后，未点继续战斗，此时强制退出游戏，默认为5秒
	st := d.status()
	if st != constant.DeskStatusCleaned &&
		st != constant.DeskStatusOver {
		if err := p.syncDeskData(); err != nil {
			log.Error(err)
		}
	} else {
		d.prepare.ready(s.UID())
		d.syncDeskStatus()
		// 必须在广播消息后调用checkStart
		d.checkStart()
	}

	return nil
}
