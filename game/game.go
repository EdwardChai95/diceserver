package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/lonng/nano"
	"github.com/lonng/nano/component"
	"github.com/lonng/nano/pipeline"
	"github.com/lonng/nano/serialize/json"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

var (
	version     = "" // 游戏版本
	forceUpdate = false
	logger      = log.WithField("component", "game")
)

func Startup() {
	rand.Seed(time.Now().Unix())
	version = viper.GetString("update.version")

	heartbeat := viper.GetInt("core.heartbeat")
	if heartbeat < 5 {
		heartbeat = 5
	}

	forceUpdate = viper.GetBool("update.force")

	logger.Infof("当前游戏服务器版本: %s, 是否强制更新: %t, 当前心跳时间间隔: %d秒", version, forceUpdate, heartbeat)
	logger.Info("game service starup")

	// register game handler
	comps := &component.Components{}
	comps.Register(defaultManager)
	comps.Register(defaultDeskManager)

	// 加密通道
	c := newCrypto()
	pip := pipeline.New()
	pip.Inbound().PushBack(c.inbound)
	pip.Outbound().PushBack(c.outbound)

	addr := fmt.Sprintf(":%d", viper.GetInt("game-server.port"))
	nano.Listen(addr,
		nano.WithPipeline(pip),
		nano.WithHeartbeatInterval(time.Duration(heartbeat)*time.Second),
		nano.WithLogger(log.WithField("component", "nano")),
		nano.WithSerializer(json.NewSerializer()),
		nano.WithComponents(comps),
	)
}
