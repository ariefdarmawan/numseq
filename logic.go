package numseq

import (
	"fmt"
	"io"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NumberStatus string

const (
	NumberStatus_Used      NumberStatus = "Used"
	NumberStatus_Blocked                = "Blocked"
	NumberStatus_Available              = "Available"
)

var (
	dh *datahub.Hub
)

func SetDataHub(h *datahub.Hub) {
	dh = h
}

func NewSequence(id string) *Sequence {
	s := new(Sequence)
	s.ID = id
	return s
}

func NewUsedSequence(sequenceid string, no int, status NumberStatus) *UsedSequence {
	us := new(UsedSequence)
	us.SequenceID = sequenceid
	us.No = no
	us.Status = string(status)
	return us
}

func Get(id string, init bool) (*Sequence, error) {
	if dh == nil {
		return nil, fmt.Errorf("method: %s, error: %s",
			"Get", "Datahub not yet initialized")
	}

	s := new(Sequence)
	e := dh.GetByID(s, id)
	if e != nil && e.Error() != "EOF" {
		fmt.Printf("Error: %s Found: %v\n", e.Error(), strings.Contains(e.Error(), "Not found"))
		e = fmt.Errorf("get fail. %s", e.Error())
	} else if e != nil {
		//fmt.Printf("Error: %s Found: %v\n", e.Error(), strings.Contains(e.Error(), "Not found"))
		if init {
			s.ID = id
			s.ReuseNumber = true
			s.LastNo = 0
			e = nil
		} else {
			e = fmt.Errorf("get fail. %s", "Not found")
		}
	}
	return s, e
}

func (s *Sequence) ChangeNumberStatus(n int, status NumberStatus) error {
	if dh == nil {
		return fmt.Errorf("method: ChangeNumberStatus, error: %s", "Datahub not yet initialized")
	}
	used := new(UsedSequence)
	if e := dh.GetByFilter(used, dbflex.And(dbflex.Eq("SequenceID", s.ID), dbflex.Eq("No", n))); e != nil {
		used = NewUsedSequence(s.ID, n, status)
	} else {
		used.Status = string(status)
	}
	e := dh.Save(used)
	if e != nil {
		return fmt.Errorf("method: %s, error: %s", "ChangeNumberStatus", e.Error())
	}
	return nil
}

func (s *Sequence) Claim() (int, error) {
	var e error
	if dh == nil {
		return 0, fmt.Errorf("method: %s, error: %s",
			"Claim", "Context not yet initialized")
	}
	var latestNo int
	latest, e := Get(s.ID, true)
	if e != nil {
		return 0, fmt.Errorf("method: %s, error: %s", "Claim",
			"Unable to get latest number - "+e.Error())
	}
	latestNo = latest.LastNo + 1

	if s.ReuseNumber {
		used := new(UsedSequence)
		e := dh.GetByParm(used,
			dbflex.NewQueryParam().
				SetSort("No").
				SetWhere(dbflex.And(dbflex.Eq("SequenceID", s.ID), dbflex.Eq("Status", NumberStatus_Available))))
		if e != nil && e != io.EOF {
			return 0, fmt.Errorf("method: %s, error: %s", "Claim",
				"Unable to get latest available - "+e.Error())
		}
		if used.No < latestNo && used.No != 0 {
			latestNo = used.No
		}
	}

	s.LastNo = latestNo
	ret := s.LastNo
	e = dh.Save(s)

	if e == nil && s.ReuseNumber {
		e = s.ChangeNumberStatus(ret, NumberStatus_Used)
	}
	return ret, e
}

func (s *Sequence) ClaimString() (string, error) {
	i, e := s.Claim()
	if e != nil {
		return "", e
	}
	if s.Format == "" {
		return fmt.Sprintf("%d", i), nil
	}
	return fmt.Sprintf(s.Format, i), nil
}

func (s *Sequence) Save() error {
	if dh == nil {
		return fmt.Errorf("method: %s, error: %s", "Save", "Context not yet initialized")
	}

	e := dh.Save(s)
	if e != nil {
		return fmt.Errorf("method: %s, error: %s", "Save", e.Error())
	}
	return nil
}

func GetNo(id string) string {
	var e error
	ns := new(Sequence)
	if e = dh.GetByID(ns, id); e != nil {
		return primitive.NewObjectID().Hex()
	}

	ret, e := ns.ClaimString()
	if e != nil {
		return primitive.NewObjectID().Hex()
	}

	return ret
}
