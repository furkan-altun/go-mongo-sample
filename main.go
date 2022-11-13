package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

var (
	Ctx             = context.TODO()
	BooksCollection *mongo.Collection
)

type Book struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id"`
	Name            string             `json:"name" bson:"name"`
	Author          string             `json:"author" bson:"author"`
	PublicationDate string             `json:"publication_date" bson:"publication_date"`
}

func main() {
	fmt.Println("Starting the api...")

	host := "127.0.0.1"
	port := "27017"

	connectionURI := "mongodb://" + host + ":" + port + "/"

	clientOptions := options.Client().ApplyURI(connectionURI)
	client, err := mongo.Connect(Ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(Ctx, nil)

	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("library")
	BooksCollection = db.Collection("books")

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/books", CreateBook).Methods("POST")
	r.HandleFunc("/api/v1/books", GetBooks).Methods("GET")
	r.HandleFunc("/api/v1/books", UpdateBook).Methods("PUT")
	r.HandleFunc("/api/v1/book/{id}", DeleteBook).Methods("DELETE")
	r.HandleFunc("/api/v1/book/{id}", GetBook).Methods("GET")

	http.ListenAndServe(":8080", r)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	author := r.FormValue("author")
	publicationDate := r.FormValue("publication_date")
	print(name, author, publicationDate)

	var book Book
	book.ID = primitive.NewObjectID()
	book.Name = name
	book.Author = author
	book.PublicationDate = publicationDate

	err := json.NewDecoder(r.Body).Decode(book)

	if err != nil {
		fmt.Println(err)
	}

	result, err := BooksCollection.InsertOne(Ctx, book)

	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(result)
}

func GetBooks(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")
	var books []Book

	cursor, err := BooksCollection.Find(Ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + err.Error() + `"}`))
	}

	defer cursor.Close(Ctx)

	for cursor.Next(Ctx) {
		var book Book
		cursor.Decode(&book)
		books = append(books, book)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + err.Error() + `"}`))
	}

	json.NewEncoder(w).Encode(books)
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
	}

	filter := bson.D{{"_id", objectId}}

	result, err := BooksCollection.DeleteOne(Ctx, filter)
	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(result)
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue("id")
	name := r.FormValue("name")
	author := r.FormValue("author")
	publicationDate := r.FormValue("publication_date")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
	}

	filter := bson.D{{"_id", objectId}}
	update := bson.D{{"$set", bson.D{{"name", name},
		{"author", author}, {"publication_date", publicationDate}}}}

	result, err := BooksCollection.UpdateOne(Ctx, filter, update)

	if err != nil {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(result)
}

func GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		fmt.Println(err)
	}

	var book Book

	err = BooksCollection.FindOne(Ctx, bson.D{{"_id", objectId}}).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	json.NewEncoder(w).Encode(book)
}
