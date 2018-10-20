package igcapi

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
TrackMongoDB stores information used to connect to the DB
*/
type TrackMongoDB struct {
	DatabaseURL         string `json:"databaseurl"`
	DatabaseName        string `json:"databasename"`
	TrackCollectionName string `json:"trackcollectionname"`
}

// Init initializes the mongo database
func (db *TrackMongoDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"trackid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

/*
Add adds a new track to the database, returns if the adding was successful
*/
func (db *TrackMongoDB) Add(t Track) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).Insert(t)
	if err != nil {
		fmt.Errorf("Error inserting track into the DB: %s", err.Error())
		return false
	}

	return true
}

/*
Count returns the amount of tracks in the database
*/
func (db *TrackMongoDB) Count() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.TrackCollectionName).Count()
	if err != nil {
		fmt.Errorf("Error retrieving the count from the database: %s", err.Error())
		return -1
	}

	return count
}

/*
Get returns the track with a given ID, and if the track was found
*/
func (db *TrackMongoDB) Get(key int) (Track, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	trackFound := true
	t := Track{}

	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).Find(bson.M{"trackid": key}).One(&t)
	if err != nil {
		trackFound = false
	}

	return t, trackFound
}
