package dice

import (
	"bytes"
	"diceserver/protocol"
	"fmt"
	"math/rand"
)

type Points []int //游戏内部表示（总共25个骰子）

func New(count int) Points {
	points := make(Points, count)
	print("len:", len(points))
	for i := range points {
		dc := rand.Intn(6) + 1
		points[i] = dc
	}
	return points
}

type Context struct {
	NumOfDice1   int
	NumOfDice2   int
	NumOfDice3   int
	NumOfDice4   int
	NumOfDice5   int
	NumOfDice6   int
	LastCallDice string

	IsLastNum bool //自己叫的点数，是否是桌面上的最后点数

	Opts   *protocol.DeskOptions
	DeskNo string
	Uid    int64
}

func (c *Context) Reset() {
	c.NumOfDice1 = -1 //真实Num从1开始
	c.NumOfDice2 = -1 //真实Num从1开始
	c.NumOfDice3 = -1 //真实Num从1开始
	c.NumOfDice4 = -1 //真实Num从1开始
	c.NumOfDice5 = -1 //真实Num从1开始
	c.NumOfDice6 = -1 //真实Num从1开始
	c.LastCallDice = ""
}

func (c *Context) String() string {
	return fmt.Sprintf("Uid=%d, DeskNo=%s, NumOfDice1=%d, NumOfDice1=%d, NumOfDice1=%d, NumOfDice1=%d, NumOfDice1=%d, NumOfDice1=%d, LastCallDice=%s, Opts=%#v",
		c.Uid, c.DeskNo, c.NumOfDice1, c.NumOfDice2, c.NumOfDice3, c.NumOfDice4,
		c.NumOfDice5, c.NumOfDice6, c.LastCallDice, c.Opts)
}

type Result []int

func (res Result) String() string {
	buf := &bytes.Buffer{}

	fmt.Fprintf(buf, "%v%v\t", PointFromIndex(res[0]), PointFromIndex(res[1]))

	for _, res := range res[2:] {
		fmt.Fprintf(buf, "%v", PointFromIndex(res))
	}
	return buf.String()
}
