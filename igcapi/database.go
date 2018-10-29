package igcapi

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
TrackDB stores information used to connect to a database storing track information
*/
type TrackDB struct {
	DatabaseURL    string `json:"databaseurl"`
	DatabaseName   string `json:"databasename"`
	CollectionName string `json:"collectionmame"`
}

/*
WebhookDB stores information used to connect to a database storing webhook information
*/
type WebhookDB struct {
	DatabaseURL    string `json:"databaseurl"`
	DatabaseName   string `json:"databasename"`
	CollectionName string `json:"collectionname"`
}

/*
Init initializes the mongo database
*/
func (db *TrackDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"tracksourceurl"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.CollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

/*
Add adds a new track to the database, returns if the adding was successful
*/
func (db *TrackDB) Add(t TrackInfo) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(t)
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

	count, err := session.DB(db.DatabaseName).C(db.CollectionName).Count()
	if err != nil {
		fmt.Printf("Error retrieving the count from the database: %s", err.Error())
		return -1
	}

	return count
}

/*
Get returns the track with a given ID, and if the track was found
*/
func (db *TrackDB) Get(key int) (TrackInfo, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	trackFound := true
	track := TrackInfo{}

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"id": key}).One(&track)
	if err != nil {
		trackFound = false
	}

	return track, trackFound
}

/*
GetAll returns all the tracks in the database, or a potential error
*/
func (db *TrackDB) GetAll() ([]TrackInfo, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	tracks := []TrackInfo{}

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{}).All(&tracks)
	if err != nil {
		return []TrackInfo{}, err
	}

	return tracks, nil
}

/*
GetAllIDs returns a slice of all the IDs used in the DB
*/
func (db *TrackDB) GetAllIDs() ([]int, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var tracks []TrackInfo

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(nil).All(&tracks)
	if err != nil {
		return []int{}, nil
	}

	IDs := []int{}
	for _, val := range tracks {
		IDs = append(IDs, val.ID)
	}

	return IDs, nil
}

/*
GetLast returns the last in the DB
*/
func (db *TrackDB) GetLast() (TrackInfo, error) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	tracks, err := db.GetAll()
	if err != nil {
		fmt.Println("Error retrieving from DB:", err.Error())
		return TrackInfo{}, err
	}
	return tracks[len(tracks)-1], nil
}

// GetLastID returns the last used track ID
func (db *TrackDB) GetLastID() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	lastTrack, err := db.GetLast()
	if err != nil {
		fmt.Println("Couldn't retrieve the last ID from the database:", err.Error())
		return -1
	}

	return lastTrack.ID
}

/*
DeleteAll deletes all tracks from the database, and returns how many tracks were deleted
*/
func (db *TrackDB) DeleteAll() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	info, err := session.DB(db.DatabaseName).C(db.CollectionName).RemoveAll(bson.M{})
	if err != nil {
		fmt.Println("Error removing from database:", err.Error())
	}

	return info.Removed
}

//
/* ------------ WebhookDB ------------ */
//

/*
Init initialises the webbook DB
*/
func (db *WebhookDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"url"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.CollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

/*
Add adds information about a webhook to the database
*/
func (db *WebhookDB) Add(wh Webhook) bool {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(wh)
	if err != nil {
		return false
	}

	return true
}

/*
GetLastID returns the last webhook ID used
*/
func (db *WebhookDB) GetLastID() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var wh Webhook

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{}).One(&wh)
	if err != nil {
		fmt.Println("Couldn't retrieve the last ID from the database.")
	}

	return wh.ID
}

/*
Get retrieves the webhook with a given ID
*/
func (db *WebhookDB) Get(ID int) Webhook {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var wh Webhook
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"id": ID}).One(&wh)
	if err != nil {
		fmt.Println("Couldn't find any Webhook with that ID")
	}

	return wh
}

/*
Delete deletes a webhook with the given ID and returns it
*/
func (db *WebhookDB) Delete(ID int) Webhook {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var wh Webhook
	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"id": ID}).One(&wh)
	if err != nil {
		fmt.Println("Couldn't find any Webhook with that ID")
	} else {
		err = session.DB(db.DatabaseName).C(db.CollectionName).Remove(bson.M{"id": ID})
	}

	return wh
}
