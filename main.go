package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"time"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

type Person struct {
	Name 	string		`json:"name,omitempty" bson:"name,omitempty"`
	Email 	string		`json:"email,omitempty" bson:"email,omitempty"`
	RSVP	string 		`json:"rsvp,omitempty" bson:"rsvp,omitempty"`
}

type Meeting struct {
	ID        		primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title 			string             `json:"title,omitempty" bson:"title,omitempty"`
	Paticipants 	[]Person           `json:"participants,omitempty" bson:"participants,omitempty"`
	Start_Time		time.Time		   `json:"start_time,omitempty" bson:"start_time,omitempty"`
	End_Time		time.Time		   `json:"end_time,omitempty" bson:"end_time,omitempty"`
	CreatedAt		time.Time		   `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

func CreateMeeting(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var meeting Meeting
	meeting.CreatedAt = time.Now()
	_ = json.NewDecoder(request.Body).Decode(&meeting)
	collection := client.Database("appointy").Collection("meetings")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, _ := collection.InsertOne(ctx, meeting)
	json.NewEncoder(response).Encode(result)
}


func GetMeeting(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var meeting Meeting
	collection := client.Database("appointy").Collection("meetings")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, Meeting{ID: id}).Decode(&meeting)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meeting)
}

func GetTimeFrameMeetings(response http.ResponseWriter, request *http.Request){
		
	start, _ := time.Parse(time.RFC3339,request.FormValue("start"))
	end, _ := time.Parse(time.RFC3339,request.FormValue("end"))
	
	filter := bson.M{
		"$and": []bson.M{
			{"start_time": bson.M{"$gte" : start}},
			{"end_time": bson.M{"$lte" : end}},
		},
	}


	//db.getCollection('meetings').find({$and:[{"start_time":{ $gte:ISODate("2020-09-19T12:00:00Z")}},{"end_time":{ $lte:ISODate("2020-09-19T18:00:00Z")}}]})

	response.Header().Set("content-type", "application/json")
	var meetings []Meeting
	collection := client.Database("appointy").Collection("meetings")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	filterCursor, err := collection.Find(ctx, filter)
	if err != nil {
    	log.Fatal(err)
	}
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer filterCursor.Close(ctx)

	for filterCursor.Next(ctx) {
		var meeting Meeting
		filterCursor.Decode(&meeting)
		meetings = append(meetings, meeting)
	}
	if err := filterCursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(meetings)
}

func main() {
	fmt.Println("Starting the application...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)

	router := mux.NewRouter()
	router.HandleFunc("/meetings", CreateMeeting).Methods("POST")
	router.HandleFunc("/meetings/{id}", GetMeeting).Methods("GET")
	router.HandleFunc("/meetings", GetTimeFrameMeetings).Methods("GET")
	http.ListenAndServe(":12345", router)
}


/*
TESTING REQUESTS

Create meeting
	http://localhost:12345/meetings
	Body:-
	{
    "title": "with Spiderman",
    "participants": [
        {
            "name": "Fedrick",
            "email": "f123.com",
            "rsvp": "Yes"
        },
        {
            "name": "Joaqium",
            "email": "joe123.com",
            "rsvp": "No"
        },
        {
            "name": "Zoe",
            "email": "zoe123.com",
            "rsvp": "Yes"
        }
    ],
    "start_time": "2020-09-19T13:00:00Z",
    "end_time": "2020-09-19T17:00:00Z"
	}

Find meeting with id
http://localhost:12345/meetings/<meetingId>	


get meetings in a particular timeframe:
http://localhost:12345/meetings?start=2020-09-19T12:00:00Z&end=2020-09-19T18:00:00Z
*/