package web

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"diceserver/db"
	"diceserver/game"
	"diceserver/web/api"

	"diceserver/pkg/errutil"
	"diceserver/protocol"
	"github.com/lonng/nex"
	log "github.com/sirupsen/logrus"
)

func authFilter(_ context.Context, r *http.Request) (context.Context, error) {
	parts := strings.Split(r.RemoteAddr, ":")
	if len(parts) < 2 {
		return context.Background(), errutil.ErrPermissionDenied
	}

	if parts[0] != "127.0.0.1" {
		return context.Background(), errutil.ErrPermissionDenied
	}

	return context.Background(), nil
}

func broadcast(query *nex.Form) (*protocol.StringMessage, error) {
	message := strings.TrimSpace(query.Get("message"))
	if message == "" || len(message) < 5 {
		return nil, errors.New("消息不可小于5个字")
	}
	api.AddMessage(message)
	game.BroadcastSystemMessage(message)
	return protocol.SuccessMessage, nil
}

func resetPlayerHandler(query *nex.Form) (*protocol.StringMessage, error) {
	uid := query.Int64OrDefault("uid", -1)
	if uid <= 0 {
		return nil, errutil.ErrIllegalParameter
	}
	log.Infof("手动重置玩家数据: Uid=%d", uid)
	game.Reset(uid)
	return protocol.SuccessMessage, nil
}

func onlineHandler(query *nex.Form) (interface{}, error) {
	begin := query.Int64OrDefault("begin", 0)
	end := query.Int64OrDefault("end", -1)
	if end < 0 {
		end = time.Now().Unix()
	}

	log.Infof("获取在线数据信息: begin=%s, end=%s", time.Unix(begin, 0).String(), time.Unix(end, 0).String())
	return db.OnlineStats(begin, end)
}

func userInfoHandler(query *nex.Form) (interface{}, error) {
	id := query.Int64OrDefault("id", -1)
	if id <= 0 {
		return nil, errutil.ErrIllegalParameter
	}

	return db.QueryUserInfo(id)
}
