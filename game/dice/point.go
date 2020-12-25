package dice

import (
	"math/rand"
	"time"
)

type Point struct {
	Id    int
	Rank  int // 点数
	Index int // 索引(1~5)
}

func PointFromIndex(idx int) *Point {
	if idx < 0 || idx > MaxPointIndex {
		return nil
	}

	return &Point{
		Rank:  idx % 6,
		Index: idx,
	}
}

func Roll() int {
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(6-1) + 1

	return idx
}

func PointFromID(id int) *Point {
	if id < 0 {
		panic("illegal point id")
	}
	return &Point{Id: id}
}
