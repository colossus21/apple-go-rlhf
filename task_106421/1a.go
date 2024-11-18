//go:build 1a
// +build 1a

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

// Data model definition

type Author struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

type Book struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Title    string             `bson:"title"`
	AuthorID primitive.ObjectID `bson:"author_id"`
	Author   Author             `bson:"author,omitempty"` // Eager loaded author
}

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")
	booksCollection := db.Collection("books")
	authorsCollection := db.Collection("authors")

	// Sample data insertion (for demonstration purposes)
	insertSampleData(authorsCollection, booksCollection)

	// Eager loading example
	var books []Book
	err = booksCollection.Aggregate(context.TODO(), []bson.M{
		{"$lookup": bson.M{
			"from":         "authors",
			"localField":   "author_id",
			"foreignField": "_id",
			"as":           "author",
		}},
		{"$unwind": "$author"},
	}).All(context.TODO(), &books)

	if err != nil {
		log.Fatal(err)
	}

	// Display the loaded books with their authors
	for _, book := range books {
		fmt.Printf("Book Title: %s, Author: %s\n", book.Title, book.Author.Name)
	}
}

func insertSampleData(authorsCollection *mongo.Collection, booksCollection *mongo.Collection) {
	// Sample authors
	authors := []Author{
		{Name: "J.K. Rowling"},
		{Name: "William Shakespeare"},
	}

	_, err := authorsCollection.InsertMany(context.TODO(), authors)
	if err != nil {
		log.Fatal("Error inserting authors:", err)
	}

	// Sample books with author references
	books := []Book{
		{Title: "Harry Potter and the Sorcerer's Stone", AuthorID: authors[0].ID},
		{Title: "Hamlet", AuthorID: authors[1].ID},
		{Title: "To Kill a Mockingbird", AuthorID: authors[1].ID},
	}

	_, err = booksCollection.InsertMany(context.TODO(), books)
	if err != nil {
		log.Fatal("Error inserting books:", err)
	}
}
