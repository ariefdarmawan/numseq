package numseq

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsedSequence struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `json:"_id" bson:"_id" key:"1"`
	SequenceID        string
	No                int
	Used              time.Time
	Status            string
}

func (u *UsedSequence) TableName() string {
	return "KNUsed"
}

func (u *UsedSequence) SetID(keys ...interface{}) {
	u.ID = keys[0].(string)
}

func (u *UsedSequence) PreSave(dbflex.IConnection) error {
	if u.ID == "" {
		u.ID = primitive.NewObjectID().Hex()
	}
	if u.SequenceID == "" || u.No == 0 || u.Status == "" {
		return errors.New("data is not complete")
	}
	if u.Status == string(NumberStatus_Used) {
		u.Used = time.Now()
	}
	return nil
}
