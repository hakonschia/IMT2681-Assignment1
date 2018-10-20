package igcapi

import (
	"reflect"
	"testing"

	"github.com/marni/goigc"

	mgo "gopkg.in/mgo.v2"
)

func setup(t *testing.T) *TrackMongoDB {
	db := &TrackMongoDB{
		DatabaseURL:         "mongodb://localhost",
		DatabaseName:        "testTrackDB",
		TrackCollectionName: "tracks",
	}

	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()

	if err != nil {
		t.Error(err)
	}

	return db
}

func tearDown(t *testing.T, db *TrackMongoDB) {
	session, err := mgo.Dial(db.DatabaseURL)
	defer session.Close()

	if err != nil {
		t.Error(err)
	}

	err = session.DB(db.DatabaseName).DropDatabase()
	if err != nil {
		t.Error(err)
	}
}

func Test_addTrackToDB(t *testing.T) {
	db := setup(t)
	defer tearDown(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Errorf("Database not initialised properly, count is %d", db.Count())
	}

	newTrack := Track{
		Track:   igc.NewTrack(),
		TrackID: 1,
	}

	db.Add(newTrack)
	if db.Count() != 1 {
		t.Errorf("Adding failed: database count expected to be 1, got %d", db.Count())
	}
}

func Test_getTrackFromDB(t *testing.T) {
	db := setup(t)
	defer tearDown(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Errorf("Database not initialised properly, count is %d", db.Count())
	}

	parsedTrack, err := igc.ParseLocation("http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc")
	if err != nil {
		t.Error("Couldn't parse the track URL")
		return
	}

	newTrack := Track{
		Track:   parsedTrack,
		TrackID: 1,
	}

	db.Add(newTrack)
	if db.Count() != 1 {
		t.Errorf("Adding failed: database count expected to be 1, got %d", db.Count())
	}

	id := 1
	trackFromDB, found := db.Get(id)
	if !found {
		t.Error("Couldn't find a track with id %D", id)
	}

	if reflect.DeepEqual(newTrack, trackFromDB) {
		t.Errorf("Tracks are not equal")
	}

}

func Test_getCountFromDB(t *testing.T) {
	db := setup(t)
	defer tearDown(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Errorf("Database not initialised properly, count is %d", db.Count())
	}
}

func Test_addDuplicateToDB(t *testing.T) {
	db := setup(t)
	defer tearDown(t, db)

	db.Init()
	if db.Count() != 0 {
		t.Errorf("Database not initialised properly, count is %d", db.Count())
	}

	parsedTrack, err := igc.ParseLocation("http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc")
	if err != nil {
		t.Error("Couldn't parse the track URL")
		return
	}

	newTrack := Track{
		Track:   parsedTrack,
		TrackID: 1,
	}

	_ = db.Add(newTrack)

	addedTwice := db.Add(newTrack)
	if addedTwice {
		t.Error("The same track could be added twice")
	}
}
