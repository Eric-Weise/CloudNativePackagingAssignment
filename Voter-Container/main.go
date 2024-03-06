package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"drexel.edu/voter/api"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Global variables to hold the command line flags to drive the todo CLI
// application
var (
	hostFlag string
	portFlag uint
)

func processCmdLineFlags() {

	flag.StringVar(&hostFlag, "h", "0.0.0.0", "Listen on all interfaces")
	flag.UintVar(&portFlag, "p", 1080, "Default Port")

	flag.Parse()
}

var rdb *redis.Client

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:63789",
	})

	processCmdLineFlags()
	r := gin.Default()
	r.Use(cors.Default())

	apiHandler, err := api.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r.GET("/voter", apiHandler.ListAllVoters)
	r.POST("/voter", apiHandler.AddVoter)
	r.PUT("/voter/:id", apiHandler.UpdateVoter)
	r.DELETE("/voter", apiHandler.DeleteAllVoters)
	r.DELETE("/voter/:id", apiHandler.DeleteVoter)
	r.GET("/voter/:id", apiHandler.GetVoter)

	r.GET("/voter/:id/polls", apiHandler.GetPollHistoryFromVoter)
	r.GET("/voter/:id/polls/:pollid", apiHandler.GetSinglePollFromVoter)
	r.POST("/voter/:id", apiHandler.AddSinglePollToVoter)

	r.GET("/health", apiHandler.HealthCheck)

	//We will now show a common way to version an API and add a new
	//version of an API handler under /v2.  This new API will support
	//a path parameter to search for todos based on a status
	// v2 := r.Group("/v2")
	// v2.GET("/voter", apiHandler.ListSelectVoters)

	serverPath := fmt.Sprintf("%s:%d", hostFlag, portFlag)
	r.Run(serverPath)
	log.Println("Starting server on ", serverPath)
	defer rdb.Close()
}
