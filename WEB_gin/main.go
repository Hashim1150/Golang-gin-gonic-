package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var NAME string
var PASSWORD string

func MongoConnection() (*mongo.Client, error) {
	uri := "mongodb://localhost:27017"
	ClientOption := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), ClientOption)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db : %v", err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping to db : %v", err)
	}
	return client, nil
}

type USers struct {
	//Uid      primitive.ObjectID `bson:"_id,omitempty" json `
	Name     string `json : "name"  bson:"name,omitempty"`
	Password string `json : "password" bson:"password,omitempty"`
}

func main() {

	router := gin.New()
	router.POST("/login", login)
	router.POST("/signup", signup)
	router.Run()

}
func login(c *gin.Context) {
	var creds USers
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}
	//cross checking

	NAME = creds.Name
	PASSWORD = creds.Password
	fmt.Println("Name :", NAME, "Password :", PASSWORD)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(NAME, PASSWORD string) { // using go routines to make it async
		c1 := c.Copy()
		//mongo connection
		defer wg.Done()

		client, err := MongoConnection()
		if err != nil {
			c1.JSON(http.StatusInternalServerError, gin.H{
				"error ": err.Error()})
			return
		}
		defer client.Disconnect(context.Background())
		collection := client.Database("mydatabase").Collection("Users")

		// check if exist ??

		var temp0 USers
		val := collection.FindOne(context.Background(), bson.M{
			"name": NAME,
		})
		err = val.Decode(&temp0)
		if err == mongo.ErrNoDocuments {
			c1.JSON(http.StatusNotFound, gin.H{
				"message": "no username found.... please signup..."})
			return
		}

		// checking for password matches for username
		var temp USers
		value := collection.FindOne(context.Background(), bson.M{
			"name":     NAME,
			"password": PASSWORD,
		})
		err = value.Decode(&temp)
		if err == mongo.ErrNoDocuments {
			c1.JSON(http.StatusNotFound, gin.H{
				"message": "wrong password"})
			return
		}
		if err != nil {
			c1.JSON(http.StatusInternalServerError, gin.H{
				"error ": err.Error(),
			})
			return
		} else {
			c1.JSON(http.StatusAccepted, gin.H{
				"message": "login succesful",
				"details": temp,
			})
		}
	}(NAME, PASSWORD)
	wg.Wait()
}

func signup(c *gin.Context) {
	var creds USers // ==>>>> for storing value from json
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error ": err.Error()})
		return
	}
	// cross check            // can skip this step
	NAME = creds.Name
	PASSWORD = creds.Password
	fmt.Println("Name :", NAME, "Password :", PASSWORD)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(name, password string) { // using go routines to make it async
		c2 := c.Copy()
		//connecting to mongodb
		defer wg.Done()

		client, err := MongoConnection()
		if err != nil {
			c2.JSON(http.StatusInternalServerError, gin.H{
				"error ": err.Error()})
			return
		}
		defer client.Disconnect(context.Background())
		collection := client.Database("mydatabase").Collection("Users")

		// <<<<<     logic here     >>>>

		var temp USers
		val := collection.FindOne(context.Background(), bson.M{
			"name": NAME,
		})
		err = val.Decode(&temp)
		if err == mongo.ErrNoDocuments {

			result, err := collection.InsertOne(context.Background(), bson.M{
				"name":     NAME,
				"password": PASSWORD,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "cant signup right now!!"})
			} else {
				c2.JSON(http.StatusCreated, gin.H{
					"message":  "signup sucessfully now please login!!",
					"Details":  result,
					"name":     NAME,
					"password": PASSWORD,
				})
			}
		} else {
			c2.JSON(http.StatusNotAcceptable, gin.H{
				"message": "user already exist !!.. please try to login.",
			})
		}
	}(NAME, PASSWORD)
	wg.Wait()
}
