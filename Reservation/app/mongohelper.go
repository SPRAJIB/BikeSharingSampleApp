package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoDB details with session, db and collection
type MongoHelper struct {
	session    *mgo.Session
	database   *mgo.Database
	collection *mgo.Collection
}

// Connect to the MongoDB
func CreateMongoConnection() *MongoHelper {
	uri := os.Getenv("MONGO_CONNECTIONSTRING")
	if uri == "" {
		uri = reservationMongoDBConnectionString
	}

	useSsl := false
	if strings.Contains(uri, "?ssl=true") {
		uri = strings.TrimSuffix(uri, "?ssl=true")
		useSsl = true
	}

	dialInfo, err := mgo.ParseURL(uri)
	if err != nil {
		fmt.Println("failed to parse URI: ", err)
		os.Exit(1)
	}

	if useSsl {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = true
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}

	var mongoSession *mgo.Session
	maxTries := 5
	for i := 1; i <= maxTries; i++ {
		mongoSession, err = mgo.DialWithInfo(dialInfo)
		if err == nil {
			break
		}

		if i < maxTries {
			fmt.Printf("%d/%d - Couldn't connect, sleeping and trying again\n", i, maxTries)
			time.Sleep(1 * time.Second)
		} else {
			fmt.Printf("%d/%d - Couldn't connect.\n", i, maxTries)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to mongodb: %s\n", err)
		os.Exit(1)
	}

	//mongoSession.SetMode(mgo.PrimaryPreferred, true)
	//mongoSession.SetSafe(&mgo.Safe{WMode: "majority"})

	db := os.Getenv("MONGO_DBNAME")
	if db == "" {
		db = reservationMongoDBDatabase
	}

	collection := os.Getenv("MONGO_COLLECTION")
	if collection == "" {
		collection = reservationMongoDBCollection
	}
	
	mongoHelper := &MongoHelper{
		session:    mongoSession,
		database:   mongoSession.DB(db),
		collection: mongoSession.DB(db).C(collection),
	}

	return mongoHelper
}

func InsertDocument(doc interface{}) error {
	mongoHelper := DbConnection
	if err := mongoHelper.collection.Insert(doc); err != nil {
		return err
	}

	return nil
}

func QueryOne(query bson.M, result interface{}) error {
	mongoHelper := DbConnection
	if err := mongoHelper.collection.Find(query).One(result); err != nil {
		return err
	}

	return nil
}

func QueryAll(query bson.M, result interface{}) error {
	mongoHelper := DbConnection
	if err := mongoHelper.collection.Find(query).All(result); err != nil {
		return err
	}

	return nil
}