package game

import (
	"diceserver/pkg/constant"
	"diceserver/pkg/errutil"
	"diceserver/pkg/room"
	"diceserver/protocol"
	"fmt"
	"strings"
	"time"

	"github.com/lonng/nano/component"
	"github.com/lonng/nano/session"
)

const (
	Offline = "离线"
	Waiting = "等待中"

	fieldDesk   = "desk"
	fieldPlayer = "player"
)

const errorCode = -1 //错误码

const (
	deskNotFoundMessage = "您加入的房间号不存在"
)

var (
	deskNotFoundResponse = &protocol.JoinDeskResponse{Code: errutil.YXDeskNotFound, Error: deskNotFoundMessage}
	deskPlayerNumEnough  = &protocol.EnterPublicDeskResponse{Code: 30001, Error: "您加入的房间已经满人"}
	reentryDesk          = &protocol.EnterPublicDeskResponse{Code: 30003, Error: "你当前正在房间中"}
)

type (
	DeskManager struct {
		component.Base
		// 桌子数据
		desks map[room.Number]*Desk // 所有桌子
	}
)

var defaultDeskManager = NewDeskManager()

func NewDeskManager() *DeskManager {
	return &DeskManager{
		desks: map[room.Number]*Desk{},
	}
}

func (manager *DeskManager) AfterInit() {
	session.Lifetime.OnClosed(func(s *session.Session) {
		// Fixed: 玩家WIFI切换到4G网络不断开, 重连时，将UID设置为illegalSessionUid
		if s.UID() > 0 {
			if err := manager.onPlayerDisconnect(s); err != nil {
				logger.Errorf("玩家退出: UID=%d, Error=%s", s.UID, err.Error())
			}
		}
	})

	// 每5分钟清空一次已摧毁的房间信息
	/*scheduler.NewTimer(300*time.Second, func() {
		destroyDesk := map[room.Number]*Desk{}
		deadline := time.Now().Add(-24 * time.Hour).Unix()
		for no, d := range manager.desks {
			// 清除创建超过24小时的房间
			if d.status() == constant.DeskStatusDestory || d.createdAt < deadline {
				destroyDesk[no] = d
			}
		}
		for _, d := range destroyDesk {
			d.destroy()
		}

		manager.dumpDeskInfo()

		// 统计结果异步写入数据库
		sCount := defaultManager.sessionCount()
		dCount := len(manager.desks)
		async.Run(func() {
			db.InsertOnline(sCount, dCount)
		})
	})*/
}

func (manager *DeskManager) dumpDeskInfo() {
	c := len(manager.desks)
	if c < 1 {
		return
	}

	logger.Infof("剩余房间数量: %d 在线人数: %d  当前时间: %s", c, defaultManager.sessionCount(), time.Now().Format("2006-01-02 15:04:05"))
	for no, d := range manager.desks {
		logger.Debugf("房号: %s, 状态: %s, 当前局数: %d",
			no, d.status().String(), d.round)
	}
}

func (manager *DeskManager) onPlayerDisconnect(s *session.Session) error {
	//uid := s.UID()
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}
	p.logger.Debug("DeskManager.onPlayerDisconnect: 玩家网络断开")

	// 移除session
	p.removeSession()

	/*if p.desk == nil || p.desk.isDestroy() {
		defaultManager.offline(uid)
		return nil
	}*/

	d := p.desk
	d.onPlayerExit(s, true)
	return nil
}

// 根据桌号返回牌桌数据
func (manager *DeskManager) desk(number room.Number) (*Desk, bool) {
	d, ok := manager.desks[number]
	return d, ok
}

// 设置桌号对应的牌桌数据
func (manager *DeskManager) setDesk(number room.Number, desk *Desk) {
	if desk == nil {
		delete(manager.desks, number)
		logger.WithField(fieldDesk, number).Debugf("清除房间: 剩余: %d", len(manager.desks))
	} else {
		manager.desks[number] = desk
	}
}

// 检查登录玩家关闭应用之前是否正在游戏
func (manager *DeskManager) UnCompleteDesk(s *session.Session, _ []byte) error {
	resp := &protocol.UnCompleteDeskResponse{}

	p, err := playerWithSession(s)
	if err != nil {
		return nil
	}
	if p.desk == nil {
		p.logger.Debug("DeskManager.UnCompleteDesk: 玩家不在房间内")
		return s.Response(resp)
	}
	d := p.desk
	if d.isRoundOver() {
		delete(manager.desks, d.roomNo)
		p.desk = nil
		p.logger.Debug("DeskManager.UnCompleteDesk: 房间已结束")
		return s.Response(resp)
	}

	return s.Response(&protocol.UnCompleteDeskResponse{
		Exist: true,
		TableInfo: protocol.TableInfo{
			DeskNo: string(d.roomNo),
			Title:  d.title(),
			Desc:   d.desc(true),
			Mode:   d.mode,
		},
	})
}

// 网络断开后, 重新连接网络
func (manager *DeskManager) ReConnect(s *session.Session, req *protocol.ReConnect) error {
	uid := req.Uid

	// 绑定UID
	if err := s.Bind(uid); err != nil {
		return err
	}

	logger.Infof("玩家重新连接服务器: UID=%d", uid)

	// 设置用户
	p, ok := defaultManager.player(uid)
	if !ok {
		logger.Infof("玩家之前用户信息已被清除，重新初始化用户信息: UID=%d", uid)
		ip := ""
		if parts := strings.Split(s.RemoteAddr().String(), ":"); len(parts) > 0 {
			ip = parts[0]
		}
		p = newPlayer(s, uid, req.Name, req.HeadUrl, ip, req.Sex)
		defaultManager.setPlayer(uid, p)
	} else {
		logger.Infof("玩家之前用户信息存在服务器上，替换session: UID=%d", uid)

		// 重置之前的session
		prevSession := p.session
		if prevSession != nil {
			prevSession.Clear()
			prevSession.Close()
		}

		// 绑定新session
		p.bindSession(s)

		// 移除广播频道
		if d := p.desk; d != nil && prevSession != nil {
			d.group.Leave(prevSession)
		}
	}

	return nil
}

// 网络断开后, 如果ReConnect后发现当前正在房间中, 则重新进入, 桌号是之前的桌号
func (manager *DeskManager) ReJoin(s *session.Session, data *protocol.ReJoinDeskRequest) error {
	/*d, ok := manager.desk(room.Number(data.DeskNo))
	if !ok || d.isDestroy() {
		return s.Response(&protocol.ReJoinDeskResponse{
			Code:  -1,
			Error: "房间已解散",
		})
	}*/
	d, ok := manager.desk(room.Number(data.DeskNo))
	if !ok || d.isRoundOver() {
		return s.Response(&protocol.ReJoinDeskResponse{
			Code:  -1,
			Error: "房间已结束",
		})
	}
	d.logger.Debugf("玩家重新加入房间: UID=%d, Data=%+v", s.UID(), data)

	return d.onPlayerReJoin(s)
}

// 应用退出后重新进入房间
func (manager *DeskManager) ReEnter(s *session.Session, msg *protocol.ReEnterDeskRequest) error {
	p, err := playerWithSession(s)
	if err != nil {
		logger.Errorf("玩家重新进入房间: UID=%d", s.UID())
		return nil
	}

	if p.desk == nil {
		p.logger.Debugf("玩家没有未完成房间，但是发送了重进请求: 请求房号: %s", msg.DeskNo)
		return nil
	}

	d := p.desk

	if string(d.roomNo) != msg.DeskNo {
		p.logger.Debugf("玩家正在试图进入非上次未完成房间: 房号: %s", d.roomNo)
		return nil
	}

	return d.onPlayerReJoin(s)
}

/*
func (manager *DeskManager) Pause(s *session.Session, _ []byte) error {
	uid := s.UID()
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}

	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}

	p.logger.Debug("玩家切换到后台")
	d.dissolve.updateOnlineStatus(uid, false)

	return nil
}

func (manager *DeskManager) Resume(s *session.Session, _ []byte) error {
	uid := s.UID()
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}

	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}

	// 玩家并未暂停
	if d.dissolve.isOnline(uid) {
		return nil
	}

	p.logger.Debug("玩家切换到前台")
	d.dissolve.updateOnlineStatus(uid, true)

	// 人数不够, 未开局, 或没有人申请解散
	if len(d.players) < d.totalPlayerCount() {
		return nil
	}

	// 有玩家切出游戏, 切回来时发现已经有人申请解散, 则刷新最新的解散状态
	p.logger.Debug("已经有人申请退出了")
	dissolveStatus := &protocol.DissolveStatusResponse{
		DissolveStatus: d.collectDissolveStatus(),
		RestTime:       d.dissolve.restTime,
	}

	return d.group.Broadcast("onDissolveStatus", dissolveStatus)
}

// 理牌结束
func (manager *DeskManager) QiPaiFinished(s *session.Session, msg []byte) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}

	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}

	return d.qiPaiFinished(s.UID())
}
*/

// Exit 处理玩家退出, 客户端会在房间人没有满的情况下发送DeskManager.Exit消息, 如果人满, 或游戏
// 开始, 客户端则发送DeskManager.Dissolve申请解散
func (manager *DeskManager) Exit(s *session.Session, msg *protocol.ExitRequest) error {
	uid := s.UID()
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}
	p.logger.Debugf("DeskManager.Exit: %+v", msg)
	d := p.desk

	if d.status() != constant.DeskStatusPlaying {
		p.logger.Debug("房间已经开始，中途不能退出")
		return nil
	}

	deskPos := -1
	for i, p := range d.players {
		if p.Uid() == uid {
			deskPos = i
			if !d.prepare.isReady(uid) {
				// fixed: 玩家在未准备的状态退出游戏, 不应该直接返回
				msg := &protocol.ExitResponse{
					AccountId: uid,
					IsExit:    true,
					ExitType:  protocol.ExitTypeExitDeskUI,
					DeskPos:   deskPos,
				}
				if err := s.Push("onDissolve", msg); err != nil {
					return err
				}
			}
			break
		}
	}

	res := &protocol.ExitResponse{
		AccountId: uid,
		IsExit:    true,
		ExitType:  protocol.ExitTypeExitDeskUI,
		DeskPos:   deskPos,
	}
	route := "onPlayerExit"
	if msg.IsDestroy {
		route = "onDissolve"
	}
	d.group.Broadcast(route, res)

	p.logger.Info("DeskManager.Exit: 退出房间")
	d.onPlayerExit(s, false)

	return nil
}

func (manager *DeskManager) OnOpen(s *session.Session, msg *protocol.OnOpenRequest) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}
	call := msg.CalledDice
	turn := msg.Turn
	beturn := msg.Beturn
	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}

	d.showCheck(turn, beturn, call)

	return nil
}

func (manager *DeskManager) OnCall(s *session.Session, msg *protocol.OnCallRequest) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}
	call := msg.CalledDice
	if call == "" {
		return fmt.Errorf("玩家选择不能为空, =%s", call)
	}
	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}
	d.showOpen(call)
	return nil
}

func (manager *DeskManager) OnGuo(s *session.Session, msg *protocol.OnGuoRequest) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}
	call := msg.CalledDice
	if call == "" {
		return fmt.Errorf("玩家选择不能为空, =%s", call)
	}
	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}
	if p.turn == d.curTurn {
		d.showCall(call)
	}
	return nil
}

func (manager *DeskManager) Ready(s *session.Session, _ []byte) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}

	d := p.desk
	d.prepare.ready(s.UID())
	println("roomNo:", d.roomNo)
	d.syncDeskStatus()
	println("Mode:", d.mode)
	// 必须在广播消息以后调用checkStart
	d.checkStart()
	return err
}

func (manager *DeskManager) ClientInitCompleted(s *session.Session, msg *protocol.ClientInitCompletedRequest) error {
	logger.Debug(msg)
	uid := s.UID()
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}
	d := p.desk
	// 客户端准备完成后加入消息广播队列
	for _, p := range d.players {
		if p.Uid() == uid {
			if p.session != s {
				p.logger.Error("DeskManager.ClientInitCompleted: Session不一致")
			}
			p.logger.Info("DeskManager.ClientInitCompleted: 玩家加入房间广播列表")
			d.group.Add(p.session)
			break
		}
	}

	// 如果不是重新进入游戏, 则同步状态到房间所有玩家
	if !msg.IsReEnter {
		d.syncDeskStatus()
	}

	return err
}

func (manager *DeskManager) RollDiceFinished(s *session.Session, _ []byte) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}

	d := p.desk
	if d == nil {
		p.logger.Debug("玩家不在房间内")
		return nil
	}

	return d.rollDiceFinished(s.UID())
}

//创建一张桌子
/*func (manager *DeskManager) CreateDesk(s *session.Session, data *protocol.CreateDeskRequest) error {
	p, err := playerWithSession(s)
	if err != nil {
		return err
	}

	if p.desk != nil {
		return s.Response(reentryDesk)
	}
	if forceUpdate && data.Version != version {
		return s.Response(createVersionExpire)
	}

	logger.Infof("牌桌选项: %#v", data.DeskOpts)

	if !verifyOptions(data.DeskOpts) {
		return errutil.ErrIllegalParameter
	}

	// 四人模式，默认可以平胡
	if data.DeskOpts.Mode == ModeFours {
		data.DeskOpts.Pinghu = true
	}

	// TODO: 测试只打一轮
	//data.DeskOpts.MaxRound = 1

	// 非俱乐部模式房卡数判定
	if data.ClubId < 0 {
		count := requireCardCount(data.DeskOpts.MaxRound)
		if p.coin < int64(count) {
			return s.Response(deskCardNotEnough)
		}

	} else {
		if db.IsBalanceEnough(data.ClubId) == false {
			return s.Response(clubCardNotEnough)
		}
	}

	no := room.Next()
	d := NewDesk(no, data.DeskOpts, data.ClubId)
	d.createdAt = time.Now().Unix()
	d.creator = s.UID()
	//房间创建者自动join
	if err := d.playerJoin(s, false); err != nil {
		return nil
	}

	// save desk information
	manager.desks[no] = d

	resp := &protocol.CreateDeskResponse{
		TableInfo: protocol.TableInfo{
			DeskNo:    string(no),
			CreatedAt: d.createdAt,
			Creator:   s.UID(),
			Title:     d.title(),
			Desc:      d.desc(true),
			Status:    d.status(),
			Round:     d.round,
			Mode:      d.opts.Mode,
		},
	}
	d.logger.Infof("当前已有牌桌数: %d", len(manager.desks))
	return s.Response(resp)
}*/

//新join在session的context中尚未有desk的cache
/*func (manager *DeskManager) Join(s *session.Session, data *protocol.JoinDeskRequest) error {
	dn := room.Number(data.DeskNo)
	d, ok := manager.desk(dn)
	if !ok {
		return s.Response(deskNotFoundResponse)
	}

	if len(d.players) >= d.totalPlayerCount() {
		return s.Response(deskPlayerNumEnough)
	}

	if err := d.playerJoin(s, false); err != nil {
		d.logger.Errorf("玩家加入房间失败，UID=%d, Error=%s", s.UID(), err.Error())
	}

	return s.Response(&protocol.JoinDeskResponse{
		TableInfo: protocol.TableInfo{
			DeskNo: d.roomNo.String(),
			Title:  d.title(),
			Desc:   d.desc(true),
			Status: d.status(),
			Mode:   d.opts.Mode,
		},
	})
}*/

//加入公共房间
func (manager *DeskManager) EnterPublicDesk(s *session.Session, data *protocol.EnterPublicDeskRequest) error {

	dn := room.Number(data.DeskNo)
	d, ok := manager.desk(dn)
	if !ok {
		d = NewDesk(dn, data.Mode)
		if err := d.playerJoin(s, false); err != nil {
			return nil
		}
		manager.desks[dn] = d
	} else {
		println("len:", len(d.players))
		println("mode:", data.Mode)
		if len(d.players) >= data.Mode {
			return s.Response(deskPlayerNumEnough)
		}
		if err := d.playerJoin(s, false); err != nil {
			d.logger.Errorf("玩家加入房间失败, UID=%d, Error=%s", s.UID(), err.Error())
		}
	}

	// 确定房间人数
	println("房间人数:", d.mode)
	resp := &protocol.EnterPublicDeskResponse{
		TableInfo: protocol.TableInfo{
			DeskNo: string(dn),
			Title:  d.title(),
			Desc:   d.desc(true),
			Status: d.status(),
			Mode:   d.mode,
		},
	}
	d.logger.Infof("当前已有牌桌数: %d", len(manager.desks))
	return s.Response(resp)
}

func (manager *DeskManager) ExitPublicDesk(s *session.Session, data *protocol.ExitPublicDeskRequest) error {

	dn := room.Number(data.DeskNo)
	d, ok := manager.desk(dn)
		if !ok {
			return s.Response(deskNotFoundResponse)
		} else {
			d.destroy()
		}

	return nil
}