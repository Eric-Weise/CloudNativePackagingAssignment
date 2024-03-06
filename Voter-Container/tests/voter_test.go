package tests

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"drexel.edu/voter/db"
	fake "github.com/brianvoe/gofakeit/v6" //aliasing package name
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

var (
	BASE_API = "http://localhost:1080"

	cli = resty.New()
)

func TestMain(m *testing.M) {

	//SETUP GOES FIRST
	rsp, err := cli.R().Delete(BASE_API + "/voter")

	if rsp.StatusCode() != 200 {
		log.Printf("error clearing database, %v", err)
		os.Exit(1)
	}

	code := m.Run()

	os.Exit(code)
}

func newRandVoter(id uint) db.Voter {
	return db.Voter{
		VoterId: id,
		Name:    fake.Name(),
		Email:   fake.Email(),
		VoteHistory: []db.VoterHistory{
			{
				PollId:   id,
				VoteId:   id,
				VoteDate: time.Time{},
			},
			{
				PollId:   id + 1,
				VoteId:   id + 1,
				VoteDate: time.Time{},
			},
		},
	}
}

func Test_LoadDB(t *testing.T) {
	numLoad := 3
	for i := 0; i < numLoad; i++ {
		item := newRandVoter(uint(i))
		rsp, err := cli.R().
			SetBody(item).
			Post(BASE_API + "/voter")

		assert.Nil(t, err)
		assert.Equal(t, 200, rsp.StatusCode())
	}
}

func Test_GetAllVoters(t *testing.T) {
	var items []db.Voter

	rsp, err := cli.R().SetResult(&items).Get(BASE_API + "/voter")

	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode())

	assert.Equal(t, 3, len(items))
}

func Test_GetOneVoter(t *testing.T) {

	var item db.Voter

	rsp, err := cli.R().SetResult(&item).Get(BASE_API + "/voter/2")
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode(), "voter #2 expected")

}

func Test_GetPollsFromVoter(t *testing.T) {

	rsp, err := cli.R().Get(BASE_API + "/voter/2/polls")
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode(), "voter #2 expected")

	var polls []db.VoterHistory
	err = json.Unmarshal(rsp.Body(), &polls)
	assert.Nil(t, err)

	assert.Greater(t, len(polls), 0, "There are no polls in this voter")

}

func Test_AddPollToVoter(t *testing.T) {
	var item db.Voter

	rsp, err := cli.R().SetResult(&item).Get(BASE_API + "/voter/2")
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode(), "voter #2 expected")

	poll := map[string]interface{}{
		"PollId":   50,
		"VoteId":   50,
		"VoteDate": time.Time{},
	}

	rsp, err = cli.R().
		SetBody(poll).
		Post(BASE_API + "/voter/2")

	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode())

}

func Test_DeleteVoter(t *testing.T) {
	var item db.Voter

	rsp, err := cli.R().SetResult(&item).Get(BASE_API + "/voter/2")
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode(), "voter #2 expected")

	rsp, err = cli.R().Delete(BASE_API + "/voter/2")
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode(), "voter not deleted expected")

	rsp, err = cli.R().SetResult(item).Get(BASE_API + "/voter/2")
	assert.Nil(t, err)
	assert.Equal(t, 404, rsp.StatusCode(), "expected not found error code")
}
