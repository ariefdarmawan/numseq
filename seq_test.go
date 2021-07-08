package numseq_test

import (
	"testing"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/numseq"
)

func prepareOrm() (*datahub.Hub, error) {
	return datahub.NewHub(datahub.GeneralDbConnBuilder("mongodb://localhost:27017/testdb"), true, 10), nil
}

func TestCreate(t *testing.T) {
	ctx, e := prepareOrm()
	if e != nil {
		t.Error(e.Error())
	}
	defer ctx.Close()

	numseq.SetDataHub(ctx)
	s, e := numseq.Get("General", true)
	if e != nil {
		t.Fatal(e)
	}
	s.Save()
}

func TestClaim(t *testing.T) {
	ctx, e := prepareOrm()
	if e != nil {
		t.Error(e.Error())
	}
	defer ctx.Close()

	numseq.SetDataHub(ctx)
	s, e := numseq.Get("General", false)
	if e != nil {
		t.Error(e.Error())
	}
	i := s.LastNo + 1

	claimed, e := s.Claim()
	if e != nil {
		t.Error(e.Error())
	}
	if i != claimed {
		t.Errorf("Error, want %d got %d", i, claimed)
	}
}

func TestClaimUsed(t *testing.T) {
	ctx, e := prepareOrm()
	if e != nil {
		t.Error(e.Error())
	}
	defer ctx.Close()

	ctx.DeleteQuery(new(numseq.UsedSequence), dbflex.Eq("SequenceID", "General"))

	numseq.SetDataHub(ctx)
	s, e := numseq.Get("General", false)
	if e != nil {
		t.Error(e.Error())
	}
	lastNo := s.LastNo

	for i := 1; i <= 5; i++ {
		s.Claim()
	}

	availNo := lastNo + 2
	s.ChangeNumberStatus(availNo, numseq.NumberStatus_Available)

	n, _ := s.Claim()
	if availNo != n {
		t.Fatalf("expect %d got %d", availNo, n)
	}
}
