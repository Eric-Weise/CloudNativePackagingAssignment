package api

import (
	"log"
	"net/http"
	"strconv"

	"drexel.edu/voter/db"
	"github.com/gin-gonic/gin"
)

// The api package creates and maintains a reference to the data handler
// this is a good design practice
type VoterAPI struct {
	db *db.VoterList
}

func New() (*VoterAPI, error) {
	dbHandler, err := db.New()
	if err != nil {
		return nil, err
	}

	return &VoterAPI{db: dbHandler}, nil
}

func (v *VoterAPI) ListAllVoters(c *gin.Context) {

	voterList, err := v.db.GetAllVoters()
	if err != nil {
		log.Println("Error Getting All Items: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if voterList == nil {
		voterList = make([]db.Voter, 0)
	}

	c.JSON(http.StatusOK, voterList)
}

func (v *VoterAPI) GetVoter(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		log.Println("Error converting id to int64: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	voter, err := v.db.GetVoter(int(id))
	if err != nil {
		log.Println("Item not found: ", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, voter)
}

func (v *VoterAPI) GetPollHistoryFromVoter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	voterHistory, err := v.db.GetVoteHistory(id)
	if err != nil {
		log.Println("Item not found:", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}
	c.JSON(http.StatusOK, voterHistory)
}

func (v *VoterAPI) GetSinglePollFromVoter(c *gin.Context) {
	voterIdStr := c.Param("id")
	pollIdStr := c.Param("pollid")

	voterid, err := strconv.Atoi(voterIdStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	pollid, err := strconv.Atoi(pollIdStr)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	poll, err := v.db.GetSingleVoteHistory(int(voterid), uint(pollid))
	if err != nil {
		log.Println("Item not found:", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}
	c.JSON(http.StatusOK, poll)
}

func (v *VoterAPI) AddSinglePollToVoter(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var poll db.VoterHistory

	if err := c.ShouldBindJSON(&poll); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if _, err := v.db.AddPoll(int(id), poll); err != nil {
		log.Println("Failed to add poll to voter:", err)
		c.AbortWithStatus(http.StatusNotFound)
	}

	c.JSON(http.StatusOK, id)
}

func (v *VoterAPI) AddVoter(c *gin.Context) {
	var voter db.Voter

	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if err := v.db.AddVoter(&voter); err != nil {
		log.Println("Error adding item: ", err)
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, voter)
}

func (v *VoterAPI) UpdateVoter(c *gin.Context) {
	var voter db.Voter
	if err := c.ShouldBindJSON(&voter); err != nil {
		log.Println("Error binding JSON: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if err := v.db.UpdateVoter(voter); err != nil {
		log.Println("Error updating voter: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	c.JSON(http.StatusOK, voter)
}

func (v *VoterAPI) DeleteVoter(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 32)

	if err := v.db.DeleteVoter(int(id)); err != nil {
		log.Println("Error deleting item: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
	}

	c.Status(http.StatusOK)
}

func (v *VoterAPI) DeleteAllVoters(c *gin.Context) {

	if err := v.db.DeleteAll(); err != nil {
		log.Println("Error deleting all items: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}

/*   SPECIAL HANDLERS FOR DEMONSTRATION - CRASH SIMULATION AND HEALTH CHECK */

func (v *VoterAPI) CrashSim(c *gin.Context) error {
	//panic() is go's version of throwing an exception
	//note with recover middleware this will not end program
	panic("Simulating an unexpected crash")
}

// implementation of GET /health. It is a good practice to build in a
// health check for your API.  Below the results are just hard coded
// but in a real API you can provide detailed information about the
// health of your API with a Health Check
func (v *VoterAPI) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"status":             "ok",
			"version":            "1.0.0",
			"uptime":             100,
			"users_processed":    1000,
			"errors_encountered": 10,
		})
}
