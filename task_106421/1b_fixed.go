//go:build 1b_fixed
// +build 1b_fixed

package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Author struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

type Book struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Title    string             `bson:"title"`
	AuthorID primitive.ObjectID `bson:"author_id"`
	Author   *Author            `bson:"author,omitempty"`
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/bookstore"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("bookstore")

	// Clear collections
	db.Collection("books").Drop(context.TODO())
	db.Collection("authors").Drop(context.TODO())

	// Sample data
	author1 := Author{ID: primitive.NewObjectID(), Name: "J.K. Rowling"}
	author2 := Author{ID: primitive.NewObjectID(), Name: "William Shakespeare"}

	authorsCollection := db.Collection("authors")
	_, err = authorsCollection.InsertOne(context.TODO(), author1)
	if err != nil {
		log.Fatal(err)
	}
	_, err = authorsCollection.InsertOne(context.TODO(), author2)
	if err != nil {
		log.Fatal(err)
	}

	book1 := Book{ID: primitive.NewObjectID(), Title: "Harry Potter and the Sorcerer's Stone", AuthorID: author1.ID}
	book2 := Book{ID: primitive.NewObjectID(), Title: "Hamlet", AuthorID: author2.ID}

	booksCollection := db.Collection("books")
	books := []interface{}{book1, book2}
	_, err = booksCollection.InsertMany(context.TODO(), books)
	if err != nil {
		log.Fatal(err)
	}

	// Eager loading using MongoDB aggregation
	booksWithAuthors, err := eagerLoadBooks(db)
	if err != nil {
		log.Fatal(err)
	}

	// Display the results
	fmt.Println("Books with Eager Loaded Authors:")
	for _, book := range booksWithAuthors {
		fmt.Printf("Book: %s, Author: %s\n", book.Title, book.Author.Name)
	}
}

func eagerLoadBooks(db *mongo.Database) ([]Book, error) {
	pipeline := []bson.M{
		{"$lookup": bson.M{
			"from":         "authors",
			"localField":   "author_id",
			"foreignField": "_id",
			"as":           "author",
		}},
		{"$unwind": "$author"},
		{"$project": bson.M{
			"_id":       1,
			"title":     1,
			"author_id": 1,
			"author":    "$author",
		}},
	}

	cursor, err := db.Collection("books").Aggregate(context.TODO(), pipeline)
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
