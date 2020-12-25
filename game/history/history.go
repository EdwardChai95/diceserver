package history

import (
	"encoding/json"
	"time"

	"diceserver/db"
	"diceserver/db/model"
	"diceserver/protocol"
)

type SnapShot struct {
	Enter     *protocol.PlayerEnterDesk `json:"enter"`
	BasicInfo *protocol.DeskBasicInfo   `json:"basicInfo"`
	Roll      *protocol.Roll            `json:"roll"`
	End       *protocol.RoundOverStats  `json:"end"`

	OpenScoreChanges []*protocol.OpenInfo `json:"openScoreChanges"`
}

type History struct {
	mode         int
	beginAt      int64
	endAt        int64
	deskID       int64
	playerName0  string
	playerName1  string
	playerName2  string
	playerName3  string
	playerName4  string
	scoreChange0 int
	scoreChange1 int
	scoreChange2 int
	scoreChange3 int
	scoreChange4 int

	SnapShot
}

func New(deskID int64, mode int, name0, name1, name2, name3, name4 string, basic *protocol.DeskBasicInfo, enter *protocol.PlayerEnterDesk, roll *protocol.Roll) *History {
	return &History{
		beginAt:     time.Now().Unix(),
		deskID:      deskID,
		mode:        mode,
		playerName0: name0,
		playerName1: name1,
		playerName2: name2,
		playerName3: name3,
		playerName4: name4,
		SnapShot: SnapShot{
			BasicInfo: basic,
			Roll:      roll,
			Enter:     enter,
		},
	}
}

func (h *History) SetEndStats(ge *protocol.RoundOverStats) error {
	h.End = ge

	return nil
}

func (h *History) SetScoreChangeForTurn(turn uint8, sc int) error {
	switch turn {
	case 0:
		h.scoreChange0 = sc
	case 1:
		h.scoreChange1 = sc
	case 2:
		h.scoreChange2 = sc
	case 3:
		h.scoreChange3 = sc
	case 4:
		h.scoreChange4 = sc
	default:
		return nil

	}
	return nil
}

func (h *History) Save() error {
	data, err := json.Marshal(&h.SnapShot)
	if err != nil {
		return err
	}

	t := &model.History{
		DeskId:       h.deskID,
		BeginAt:      h.beginAt,
		Mode:         h.mode,
		EndAt:        time.Now().Unix(),
		PlayerName0:  h.playerName0,
		PlayerName1:  h.playerName1,
		PlayerName2:  h.playerName2,
		PlayerName3:  h.playerName3,
		ScoreChange0: h.scoreChange0,
		ScoreChange1: h.scoreChange1,
		ScoreChange2: h.scoreChange2,
		ScoreChange3: h.scoreChange3,
		Snapshot:     string(data),
	}

	return db.InsertHistory(t)
}

type Record struct {
	TotalScore int `json:"totalScore"`
}

type MatchStats map[int64][]*Record

func (ps MatchStats) Result() map[int64]*Record {
	ret := make(map[int64]*Record)

	for p, records := range ps {
		if _, ok := ret[p]; !ok {
			ret[p] = &Record{}
		}

		for _, d := range records {
			ret[p].TotalScore += d.TotalScore
		}
	}
	return ret
}
