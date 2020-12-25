package game

import "diceserver/protocol"

func BroadcastSystemMessage(message string) {
	defaultManager.group.Broadcast("onBroadcast", &protocol.StringMessage{Message: message})
}

func Reset(uid int64) {
	defaultManager.chReset <- uid
}
