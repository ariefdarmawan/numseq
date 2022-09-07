package numseq

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Engine struct {
	conn dbflex.IConnection
}

func New(conn dbflex.IConnection) *Engine {
	eng := new(Engine)
	eng.conn = conn
	return eng
}

func (eng *Engine) GetOrCreate(id string, init bool, format string) (*Sequence, error) {
	if dh == nil {
		return nil, fmt.Errorf("method: %s, error: %s",
			"Get", "Datahub not yet initialized")
	}

	s := new(Sequence)
	s.ID = id
	e := orm.Get(eng.conn, s)
	if e != nil && e.Error() != "EOF" {
		e = fmt.Errorf("get fail. %s", e.Error())
	} else if e != nil {
		//fmt.Printf("Error: %s Found: %v\n", e.Error(), strings.Contains(e.Error(), "Not found"))
		if init {
			s.ID = id
			s.Format = format
			s.ReuseNumber = true
			s.LastNo = 0
			e = orm.Save(eng.conn, s)
		} else {
			e = fmt.Errorf("get fail. %s", "Not found")
		}
	}
	return s, e
}

func (eng *Engine) ChangeNumberStatus(id string, n int, status NumberStatus) error {
	conn := eng.conn
	if conn == nil {
		return fmt.Errorf("method: ChangeNumberStatus, error: %s", "Datahub not yet initialized")
	}
	used := new(UsedSequence)
	if e := orm.GetWhere(conn, used, dbflex.And(dbflex.Eq("SequenceID", id), dbflex.Eq("No", n))); e != nil {
		used = NewUsedSequence(id, n, status)
	} else {
		used.Status = string(status)
	}
	e := orm.Save(conn, used)
	if e != nil {
		return fmt.Errorf("method: %s, error: %s", "ChangeNumberStatus", e.Error())
	}
	return nil
}

func (eng *Engine) Claim(id string) (int, string, error) {
	var e error
	conn := eng.conn
	if conn == nil {
		return 0, "", fmt.Errorf("method: %s, error: %s",
			"Claim", "Context not yet initialized")
	}
	var latestNo int
	latest := new(Sequence)
	latest.ID = id
	if e = orm.Get(conn, latest); e != nil {
		return 0, "", fmt.Errorf("method: %s, error: %s", "Claim",
			"Unable to get latest number - "+e.Error())
	}
	latestNo = latest.LastNo + 1

	if latest.ReuseNumber {
		qp := dbflex.NewQueryParam().
			SetSort("No").
			SetTake(1).
			SetWhere(dbflex.And(dbflex.Eq("SequenceID", id), dbflex.Eq("Status", NumberStatus_Available)))
		useds := []UsedSequence{}
		e := orm.Gets(conn, new(UsedSequence), &useds, qp)
		if e != nil {
			return 0, "", fmt.Errorf("method: %s, error: %s", "Claim",
				"Unable to get latest available - "+e.Error())
		}
		if len(useds) > 0 && useds[0].No < latestNo {
			latestNo = useds[0].No
		}
	}

	latest.LastNo = latestNo
	ret := latest.LastNo
	e = orm.Save(conn, latest)

	if e == nil && latest.ReuseNumber {
		e = eng.ChangeNumberStatus(id, ret, NumberStatus_Used)
	}
	return ret, latest.Format, e
}

func (eng *Engine) ClaimString(id string, date *time.Time) (string, error) {
	i, format, e := eng.Claim(id)

	if e != nil {
		return "", e
	}

	if format == "" {
		return fmt.Sprintf("%d", i), nil
	}

	if date != nil {
		format = strings.ReplaceAll(format, "$year", strconv.Itoa(date.Year()))
		format = strings.ReplaceAll(format, "$month", codekit.Date2String(*date, "MM"))
	}

	return fmt.Sprintf(format, i), nil
}

func (eng *Engine) GetNo(id string, date *time.Time) string {
	var e error
	conn := eng.conn
	ns := new(Sequence)
	ns.ID = id
	if e = orm.Get(conn, ns); e != nil {
		return primitive.NewObjectID().Hex()
	}

	ret, e := eng.ClaimString(id, date)
	if e != nil {
		return primitive.NewObjectID().Hex()
	}

	return ret
}
