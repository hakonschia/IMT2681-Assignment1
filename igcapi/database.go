package igcapi

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
TrackDB stores information used to connect to the DB
*/
type TrackDB struct {
	DatabaseURL         string `json:"databaseurl"`
	DatabaseName        string `json:"databasename"`
	TrackCollectionName string `json:"trackcollectionname"`
}

// Init initializes the mongo database
func (db *TrackDB) Init() {
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
func (db *TrackDB) Add(t Track) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).Insert(t)
	if err != nil {
		fmt.Printf("Error inserting track into the DB: %s", err.Error())
		return false
	}

	return true
}

/*
Count returns the amount of tracks in the database
*/
func (db *TrackDB) Count() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.TrackCollectionName).Count()
	if err != nil {
		fmt.Printf("Error retrieving the count from the database: %s", err.Error())
		return -1
	}

	return count
}

/*
Get returns the track with a given ID, and if the track was found
*/
func (db *TrackDB) Get(key int) (Track, bool) {
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

// GetAll returns all the tracks in the database, or a potential error
func (db *TrackDB) GetAll() ([]Track, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	tracks := []Track{}

	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).Find(bson.M{}).All(&tracks)
	if err != nil {
		return []Track{}, err
	}

	return tracks, nil
}

// GetAllIDs returns a slice of all the IDs used in the DB
func (db *TrackDB) GetAllIDs() ([]int, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var tracks []Track

	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).Find(nil).All(&tracks)
	if err != nil {
		return []int{}, nil
	}

	IDs := []int{}
	for _, val := range tracks {
		IDs = append(IDs, val.TrackID)
	}

	return IDs, nil
}

// GetLastID returns the last used track ID
func (db *TrackDB) GetLastID() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var track Track
	// MongoDB sorts based on insertion time, so the last element can be found via the number of elements
	err = session.DB(db.DatabaseName).C(db.TrackCollectionName).Find(bson.M{"trackid": db.Count() - 1}).One(&track)
	fmt.Println("NextID:", track.TrackID)
	return track.TrackID
}
