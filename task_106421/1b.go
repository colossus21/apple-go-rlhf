//go:build 1b
// +build 1b

package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-driver/bson"
	"github.com/mongodb/mongo-driver/mongo/options"
	"log"
	"mongo-eager-loading/mongo"
)

// a) Data model definition

type Author struct {
	ID   bson.ObjectID `bson:"_id,omitempty"`
	Name string        `bson:"name"`
}

type Book struct {
	ID       bson.ObjectID `bson:"_id,omitempty"`
	Title    string        `bson:"title"`
	AuthorID bson.ObjectID `bson:"author_id"`
	Author   *Author       `bson:",inline"`
}

func main() {
	// Setup MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/bookstore"))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.TODO())

	// Clear the databases for a clean run
	mongo.ClearDatabases(client)

	// Sample data
	author1 := Author{Name: "J.K. Rowling"}
	author2 := Author{Name: "William Shakespeare"}

	err = mongo.InsertOne(client, "authors", author1)
	if err != nil {
		log.Fatal(err)
	}

	err = mongo.InsertOne(client, "authors", author2)
	if err != nil {
		log.Fatal(err)
	}

	book1 := Book{Title: "Harry Potter and the Sorcerer's Stone", AuthorID: author1.ID}
	book2 := Book{Title: "Hamlet", AuthorID: author2.ID}

	err = mongo.InsertMany(client, "books", []interface{}{book1, book2})
	if err != nil {
		log.Fatal(err)
	}

	// b) Eager loading using MongoDB aggregation
	booksWithAuthors, err := eagerLoadBooks(client)
	if err != nil {
		log.Fatal(err)
	}

	// Display the results
	fmt.Println("Books with Eager Loaded Authors:")
	for _, book := range booksWithAuthors {
		fmt.Printf("Book: %s, Author: %s\n", book.Title, book.Author.Name)
	}
}

func eagerLoadBooks(client *mongo.Client) ([]Book, error) {
	booksCollection := client.Database("bookstore").Collection("books")
	authorsCollection := client.Database("bookstore").Collection("authors")

	pipeline := []bson.M{
		{{"$lookup", bson.M{
			"from":         "authors",
			"localField":   "author_id",
			"foreignField": "_id",
			"as":           "author",
		}}},
		{{"$unwind", "$author"}},
		{{"$project", bson.M{
			"_id":        1,
			"title":      1,
			"author_id":  1,
			"author":     "$author",
			"_author_id": 0, // Remove the embedded _id field of the author
		}}},
	}

	cursor, err := booksCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	var booksWithAuthors []Book
	err = cursor.All(context.TODO(), &booksWithAuthors)
	if err != nil {
		return nil, err
	}

	return booksWithAuthors, nil
}
