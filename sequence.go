package numseq

import "git.kanosolution.net/kano/dbflex/orm"

type Sequence struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `json:"_id" bson:"_id" key:"1"`
	Title             string
	LastNo            int
	ReuseNumber       bool
	Format            string
}

func (s *Sequence) TableName() string {
	return "KNSequences"
}

func (s *Sequence) SetID(keys ...interface{}) {
	s.ID = keys[0].(string)
}
