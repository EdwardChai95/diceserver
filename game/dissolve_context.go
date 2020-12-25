package game

import "diceserver/protocol"

// 退出房间统计
type dissolveContext struct {
	desk   *Desk            // 桌子
	status map[int64]bool   // 退出统计
	desc   map[int64]string // 退出描述
	pause  map[int64]bool   // 离线状态
}

func newDissolveContext(desk *Desk) *dissolveContext {
	return &dissolveContext{
		desk:   desk,
		status: map[int64]bool{},
		desc:   map[int64]string{},
		pause:  map[int64]bool{},
	}
}

func (d *dissolveContext) isOnline(uid int64) bool {
	return !d.pause[uid]
}

func (d *dissolveContext) updateOnlineStatus(uid int64, online bool) {
	if online {
		delete(d.pause, uid)
	} else {
		d.pause[uid] = true
	}

	d.desk.logger.Debugf("玩家在线状态：%+v", d.pause)
	d.desk.group.Broadcast("onPlayerOfflineStatus", &protocol.PlayerOfflineStatus{Uid: uid, Offline: !online})
}

func (d *dissolveContext) offlineCount() int {
	return len(d.pause)
}
