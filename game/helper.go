package game

import (
	"diceserver/pkg/errutil"
	"diceserver/protocol"
	"github.com/lonng/nano/session"
	"runtime"
	"strings"
)

const (
	ModeTwos  = 2 // 两人模式
	ModeTrios = 3 // 三人模式
	ModeFours = 4 // 四人模式
	ModeFives = 5 // 五人模式
)

func verifyOptions(opts *protocol.DeskOptions) bool {
	if opts == nil {
		return false
	}

	if opts.Mode != ModeTwos && opts.Mode != ModeTrios && opts.Mode != ModeFours && opts.Mode != ModeFives {
		return false
	}

	return true
}

func playerWithSession(s *session.Session) (*Player, error) {
	p, ok := s.Value(kCurPlayer).(*Player)
	if !ok {
		return nil, errutil.ErrPlayerNotFound
	}
	return p, nil
}

func stack() string {
	buf := make([]byte, 10000)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)

	// skip nano frame lines
	const skip = 7
	count := 0
	index := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == skip
	})
	return s[index+1:]
}
