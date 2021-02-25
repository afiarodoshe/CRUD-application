package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// User is a struct that represents a user in our application
type User struct {
	FullName string `json:"fullName" bson:"fullName"`
	UserName string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
	Status   string `json:"status" bson:"status"`
}

// Post is a struct that represents a single post
type Post struct {
	Teacher User `json:"teacher" bson:"teacher"`
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
var posts []Post = []Post{}
var client *mongo.Client
var err error

func main() {

	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	//defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")
	fmt.Println(databases)

	router := mux.NewRouter()

	router.HandleFunc("/posts", addTeacher).Methods("POST")

	router.HandleFunc("/posts", getAllTeacherInfo).Methods("GET")

	router.HandleFunc("/posts/{email}", getTeacher).Methods("GET")

	router.HandleFunc("/posts", updateTeacherInfo).Methods("PUT")

	router.HandleFunc("/posts/{email}", deleteTeacherInfo).Methods("DELETE")

	http.ListenAndServe(":5000", router)

}

func addTeacher(w http.ResponseWriter, r *http.Request) {
	// get Item value from the JSON body
	w.Header().Set("Content-Type", "application/json")
	reqBody, _ := ioutil.ReadAll(r.Body)
	user := User{}
	err := json.Unmarshal(reqBody, &user)
	if err != nil {
		fmt.Println("Error-", err)
	}

	var emailParam string = user.Email
	user.Status = "V"
	var result bool = isEmailValid(emailParam)
	if result == true {
		var user1 bson.M
		//var emailParam string = user.Email
		var status = bson.D{{Key: "status", Value: "D"}}
		if len(status) > 0 {
			database := client.Database("teachers")
			userCollection := database.Collection("teachers")
			err = userCollection.FindOne(context.TODO(), bson.M{"email": emailParam, "status": "V"}).Decode(&user1)
			if len(user1) < 1 {
				fmt.Println("printing user 1")
				fmt.Println(user)
				fmt.Println(r.Body)
				fmt.Println(user.FullName + " " + user.Email)
				collection := client.Database("teachers").Collection("teachers")
				if _, err := collection.InsertOne(context.TODO(), bson.D{
					{Key: "fullName", Value: user.FullName},
					{Key: "username", Value: user.UserName},
					{Key: "email", Value: user.Email},
					{Key: "status", Value: "V"},
				}); err != nil {
					log.Fatal(err)
				}

				formattedData, _ := json.MarshalIndent(user, "", "   ")
				fmt.Fprintf(w, string(formattedData))

				fmt.Println("printing user")
				fmt.Print(user)
			} else {
				fmt.Println("user already exists", err)
				fmt.Fprintf(w, "user already exists")
				formattedData, _ := json.MarshalIndent(user, "", "   ")
				fmt.Fprintf(w, string(formattedData))
			}
		}
	} else {
		fmt.Printf("invalid email.. 1")
		fmt.Fprintf(w, "invalid email.. ")
	}
}
func getTeacher(w http.ResponseWriter, r *http.Request) {
	// get the Email of the post from the route parameter
	w.Header().Set("Content-Type", "application/json")
	var user bson.M
	var emailParam string = mux.Vars(r)["email"]
	var status = bson.D{{Key: "status", Value: "D"}}
	if len(status) > 0 {
		database := client.Database("teachers")
		userCollection := database.Collection("teachers")
		err := userCollection.FindOne(context.TODO(), bson.M{"email": emailParam, "status": "V"}).Decode(&user)
		if err != nil {
			// there was an error
			w.WriteHeader(400)
			w.Write([]byte("ID could not be converted to integer"))
		}
		formattedData, err := json.MarshalIndent(user, "", "   ")
		fmt.Fprintf(w, string(formattedData))
		fmt.Println(user)
	} else {
		fmt.Fprintf(w, "user deleted")
	}
}

func getAllTeacherInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user bson.M
	database := client.Database("teachers")
	userCollection := database.Collection("teachers")
	filterCursor, err := userCollection.Find(context.TODO(), bson.M{"status": "V"})
	if err != nil {
		log.Fatal(err)
	}
	if err = filterCursor.All(context.TODO(), &User{}); err != nil {
		log.Fatal(err)
	}
	formattedData, err := json.MarshalIndent(user, "", "   ")
	fmt.Fprintf(w, string(formattedData))
	fmt.Println(user)
}

func updateTeacherInfo(w http.ResponseWriter, r *http.Request) {
	// update the teacher info
	w.Header().Set("Content-Type", "application/json")
	var user bson.M
	reqBody, err := ioutil.ReadAll(r.Body)
	user1 := User{}
	err = json.Unmarshal(reqBody, &user1)
	if err != nil {
		fmt.Println("Error-", err)
	}

	// DB Connection
	database := client.Database("teachers")
	userCollection := database.Collection("teachers")
	err = userCollection.FindOne(context.TODO(), bson.M{"email": user1.Email, "status": "D"}).Decode(&user)
	fmt.Println(user)
	if len(user) != 0 {
		fmt.Fprintf(w, "user already deleted")
	} else {
		err = userCollection.FindOne(context.TODO(), bson.M{"email": user1.Email, "status": "V"}).Decode(&user)
		result, err := userCollection.UpdateOne(
			context.TODO(),
			bson.M{"email": user1.Email},
			bson.D{
				{"$set", bson.D{{"fullName", user1.FullName}}},
				{"$set", bson.D{{"username", user1.UserName}}},
				{"$set", bson.D{{"status", "V"}}},
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		formattedData, _ := json.MarshalIndent(user, "", "   ")
		fmt.Fprintf(w, string(formattedData))
		//fmt.Println(emailParam)
		fmt.Printf("ModifiedOne removed %v document(s)\n", result.ModifiedCount)
	}
}

func deleteTeacherInfo(w http.ResponseWriter, r *http.Request) {
	// get the ID of the post from the route parameters
	w.Header().Set("Content-Type", "application/json")
	var user bson.M
	var emailParam string = mux.Vars(r)["email"]
	reqBody, err := ioutil.ReadAll(r.Body)
	user1 := User{}
	err = json.Unmarshal(reqBody, &user1)
	if err != nil {
		fmt.Println("Error-", err)
	}
	database := client.Database("teachers")
	userCollection := database.Collection("teachers")
	err = userCollection.FindOne(context.TODO(), bson.M{"email": user1.Email, "status": "V"}).Decode(&user)
	result, err := userCollection.UpdateOne(
		context.TODO(),
		bson.M{"email": user1.Email},
		bson.D{
			{"$set", bson.D{{"fullName", user1.FullName}}},
			{"$set", bson.D{{"username", user1.UserName}}},
			{"$set", bson.D{{"status", "D"}}},
		},
	)
	fmt.Fprintf(w, "user deleted ..")
	formattedData, _ := json.MarshalIndent(user, "", "   ")
	fmt.Fprintf(w, string(formattedData))
	fmt.Println(emailParam)
	fmt.Printf("DeleteOne removed %v document(s)\n", result.ModifiedCount)
}
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	if !emailRegex.MatchString(e) {
		return false
	}
	parts := strings.Split(e, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}
