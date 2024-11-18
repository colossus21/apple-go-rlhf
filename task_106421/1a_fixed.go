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
}

type BookWithAuthor struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Title    string             `bson:"title"`
	AuthorID primitive.ObjectID `bson:"author_id"`
	Author   Author             `bson:"author"`
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")
	booksCollection := db.Collection("books")
	authorsCollection := db.Collection("authors")

	authors := []Author{
		{ID: primitive.NewObjectID(), Name: "J.K. Rowling"},
		{ID: primitive.NewObjectID(), Name: "William Shakespeare"},
	}

	authorsI := make([]interface{}, len(authors))
	for i, v := range authors {
		authorsI[i] = v
	}
	_, err = authorsCollection.InsertMany(context.TODO(), authorsI)
	if err != nil {
		log.Fatal("Error inserting authors:", err)
	}

	books := []Book{
		{ID: primitive.NewObjectID(), Title: "Harry Potter and the Sorcerer's Stone", AuthorID: authors[0].ID},
		{ID: primitive.NewObjectID(), Title: "Hamlet", AuthorID: authors[1].ID},
		{ID: primitive.NewObjectID(), Title: "To Kill a Mockingbird", AuthorID: authors[1].ID},
	}

	booksI := make([]interface{}, len(books))
	for i, v := range books {
		booksI[i] = v
	}
	_, err = booksCollection.InsertMany(context.TODO(), booksI)
	if err != nil {
		log.Fatal("Error inserting books:", err)
	}

	cursor, err := booksCollection.Aggregate(context.TODO(), []bson.M{
		{"$lookup": bson.M{
			"from":         "authors",
			"localField":   "author_id",
			"foreignField": "_id",
			"as":           "author",
		}},
		{"$unwind": "$author"},
	})
	if err != nil {
		log.Fatal(err)
	}

	var booksWithAuthors []BookWithAuthor
	if err = cursor.All(context.TODO(), &booksWithAuthors); err != nil {
		log.Fatal(err)
	}

	for _, book := range booksWithAuthors {
		fmt.Printf("Book Title: %s, Author: %s\n", book.Title, book.Author.Name)
	}
}
