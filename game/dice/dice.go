package dice

import (
	"math/rand"
	"sort"
	"time"
)

//根据人数分配骰子
// 1: 玩家1
// 2： 玩家2
// 3： 玩家3
// 4： 玩家4
// 5： 玩家5
const MaxPointIndex = 25

type Dice []*Point

func (d Dice) Len() int {
	return len(d)
}

func (d Dice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d Dice) Less(i, j int) bool {
	return d[i].Id < d[j].Id
}

func (d Dice) Shuffle() {
	s := rand.New(rand.NewSource(time.Now().Unix()))
	for i := range d {
		j := s.Intn(len(d))
		d[i], d[j] = d[j], d[i]
	}
}

func (d Dice) Sort() {
	sort.Sort(d)
}

func (d Dice) Indexes() []int {
	idx := make([]int, len(d))
	for i, p := range d {
		idx[i] = p.Index
	}
	return idx
}

func (d Dice) Ids() []int {
	ids := make([]int, len(d))
	for i, p := range d {
		ids[i] = p.Id
	}
	return ids
}

func FromID(ids []int) Dice {
	dc := make(Dice, len(ids))
	for i, idx := range ids {
		dc[i] = PointFromID(idx)
	}
	return dc
}
