package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nitishm/go-rejson/v4"
	"github.com/redis/go-redis/v9"
)

type VoterHistory struct {
	PollId   uint      `json:"PollId"`
	VoteId   uint      `json:"VoteId"`
	VoteDate time.Time `json:"VoteDate"`
}

type Voter struct {
	VoterId     uint           `json:"VoterId"`
	Name        string         `json:"Name"`
	Email       string         `json:"Email"`
	VoteHistory []VoterHistory `json:"VoteHistory"`
}

const (
	RedisNilError        = "redis: nil"
	RedisDefaultLocation = "0.0.0.0:6379"
	RedisKeyPrefix       = "voter:"
)

type cache struct {
	cacheClient *redis.Client
	jsonHelper  *rejson.Handler
	context     context.Context
}

// ToDo is the struct that represents the main object of our
// todo app.  It contains a reference to a cache object
type VoterList struct {
	//more things would be included in a real implementation

	//Redis cache connections
	cache
}

func New() (*VoterList, error) {
	//We will use an override if the REDIS_URL is provided as an environment
	//variable, which is the preferred way to wire up a docker container
	redisUrl := os.Getenv("REDIS_URL")
	//This handles the default condition
	if redisUrl == "" {
		redisUrl = RedisDefaultLocation
	}
	log.Println("DEBUG:  USING REDIS URL: " + redisUrl)
	return NewWithCacheInstance(redisUrl)
}

// NewWithCacheInstance is a constructor function that returns a pointer to a new
// ToDo struct.  It accepts a string that represents the location of the redis
// cache.
func NewWithCacheInstance(location string) (*VoterList, error) {

	//Connect to redis.  Other options can be provided, but the
	//defaults are OK
	client := redis.NewClient(&redis.Options{
		Addr: location,
	})

	//We use this context to coordinate betwen our go code and
	//the redis operaitons
	ctx := context.Background()

	//This is the reccomended way to ensure that our redis connection
	//is working
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Println("Error connecting to redis" + err.Error() + "cache might not be available, continuing...")
	}

	//By default, redis manages keys and values, where the values
	//are either strings, sets, maps, etc.  Redis has an extension
	//module called ReJSON that allows us to store JSON objects
	//however, we need a companion library in order to work with it
	//Below we create an instance of the JSON helper and associate
	//it with our redis connnection
	jsonHelper := rejson.NewReJSONHandler()
	jsonHelper.SetGoRedisClientWithContext(ctx, client)

	//Return a pointer to a new ToDo struct
	return &VoterList{
		cache: cache{
			cacheClient: client,
			jsonHelper:  jsonHelper,
			context:     ctx,
		},
	}, nil
}

//------------------------------------------------------------
// REDIS HELPERS
//------------------------------------------------------------

// We will use this later, you can ignore for now
func isRedisNilError(err error) bool {
	return errors.Is(err, redis.Nil) || err.Error() == RedisNilError
}

// In redis, our keys will be strings, they will look like
// todo:<number>.  This function will take an integer and
// return a string that can be used as a key in redis
func redisKeyFromId(id int) string {
	return fmt.Sprintf("%s%d", RedisKeyPrefix, id)
}

// Helper to return a ToDoItem from redis provided a key
func (v *VoterList) getItemFromRedis(key string, voter *Voter) error {

	//Lets query redis for the item, note we can return parts of the
	//json structure, the second parameter "." means return the entire
	//json structure
	voterObject, err := v.jsonHelper.JSONGet(key, ".")
	if err != nil {
		return err
	}

	//JSONGet returns an "any" object, or empty interface,
	//we need to convert it to a byte array, which is the
	//underlying type of the object, then we can unmarshal
	//it into our ToDoItem struct
	err = json.Unmarshal(voterObject.([]byte), voter)
	if err != nil {
		return err
	}

	return nil
}

func (v *VoterList) AddVoter(voter *Voter) error {

	//Before we add an item to the DB, lets make sure
	//it does not exist, if it does, return an error
	redisKey := redisKeyFromId(int(voter.VoterId))
	var existingVoter Voter
	if err := v.getItemFromRedis(redisKey, &existingVoter); err == nil {
		return errors.New("voter already exists")
	}

	//Add item to database with JSON Set
	if _, err := v.jsonHelper.JSONSet(redisKey, ".", voter); err != nil {
		return err
	}

	//If everything is ok, return nil for the error
	return nil
}

func (v *VoterList) DeleteVoter(id int) error {

	pattern := redisKeyFromId(int(id))
	numDeleted, err := v.cacheClient.Del(v.context, pattern).Result()
	if err != nil {
		return err
	}
	if numDeleted == 0 {
		return errors.New("attempted to delete non-existent item")
	}

	return nil
}

func (v *VoterList) DeleteAll() error {

	pattern := RedisKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()

	numDeleted, err := v.cacheClient.Del(v.context, ks...).Result()
	if err != nil {
		return err
	}

	if numDeleted != int64(len(ks)) {
		return errors.New("one or more items could not be deleted")
	}

	return nil
}

func (v *VoterList) UpdateVoter(voter Voter) error {

	redisKey := redisKeyFromId(int(voter.VoterId))
	var existingItem Voter
	if err := v.getItemFromRedis(redisKey, &existingItem); err != nil {
		return errors.New("item does not exist")
	}

	if _, err := v.jsonHelper.JSONSet(redisKey, ".", voter); err != nil {
		return err
	}

	return nil
}

func (v *VoterList) GetVoter(id int) (Voter, error) {

	var voter Voter
	pattern := redisKeyFromId(int(id))
	err := v.getItemFromRedis(pattern, &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

func (v *VoterList) GetAllVoters() ([]Voter, error) {

	var voterList []Voter
	var voter Voter

	pattern := RedisKeyPrefix + "*"
	ks, _ := v.cacheClient.Keys(v.context, pattern).Result()
	for _, key := range ks {
		err := v.getItemFromRedis(key, &voter)
		if err != nil {
			return nil, err
		}
		voterList = append(voterList, voter)
	}

	return voterList, nil
}

func (v *VoterList) PrintItem(voter Voter) {
	jsonBytes, _ := json.MarshalIndent(voter, "", "  ")
	fmt.Println(string(jsonBytes))
}

func (v *VoterList) PrintAllItems(voterList []Voter) {
	for _, voter := range voterList {
		v.PrintItem(voter)
	}
}

func (v *VoterList) JsonToItem(jsonString string) (Voter, error) {
	var voter Voter
	err := json.Unmarshal([]byte(jsonString), &voter)
	if err != nil {
		return Voter{}, err
	}

	return voter, nil
}

func (v *VoterList) GetVoteHistory(id int) ([]VoterHistory, error) {

	redisKey := redisKeyFromId(id)
	var existingVoter Voter
	if err := v.getItemFromRedis(redisKey, &existingVoter); err != nil {
		return existingVoter.VoteHistory, errors.New("voter does not exist")
	}

	return existingVoter.VoteHistory, nil
}

func (v *VoterList) GetSingleVoteHistory(voterId int, pollId uint) (*VoterHistory, error) {

	redisKey := redisKeyFromId(voterId)
	var existingVoter Voter
	if err := v.getItemFromRedis(redisKey, &existingVoter); err != nil {
		return nil, errors.New("voter does not exist")
	}

	for _, vote := range existingVoter.VoteHistory {
		if vote.PollId == pollId {
			return &vote, nil
		}
	}

	return nil, errors.New("poll does not exist for the specified voter")
}

func (v *VoterList) AddPoll(voterId int, poll VoterHistory) (Voter, error) {

	redisKey := redisKeyFromId(voterId)
	var existingVoter Voter
	if err := v.getItemFromRedis(redisKey, &existingVoter); err != nil {
		return existingVoter, errors.New("voter does not exist")
	}

	existingVoter.VoteHistory = append(existingVoter.VoteHistory, poll)

	if _, err := v.jsonHelper.JSONSet(redisKey, ".", existingVoter); err != nil {
		return existingVoter, err
	}

	return existingVoter, nil

}
