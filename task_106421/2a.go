//go:build 2a
// +build 2a

package main

import (
	"context"
	"fmt"
	"time"

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
	author   *Author
	cache    map[primitive.ObjectID]*Author
}

// GetAuthor performs lazy loading of the author using a database query
func (b *Book) GetAuthor(client *mongo.Client, db *mongo.Database) {
	if b.author != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authorsCollection := db.Collection("authors")
	cursor, err := authorsCollection.FindOne(ctx, bson.M{"_id": b.AuthorID})
	if err != nil {
		log.Fatal("Error finding author:", err)
	}

	var author Author
	err = cursor.Decode(&author)
	if err != nil {
		log.Fatal("Error decoding author:", err)
	}

	b.author = &author
}

// PreloadAuthors performs batch loading of authors for a given list of books
func PreloadAuthors(client *mongo.Client, db *mongo.Database, books []Book) {
	if len(books) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authorIDs := make([]primitive.ObjectID, len(books))
	for i, book := range books {
		authorIDs[i] = book.AuthorID
	}

	authorsCollection := db.Collection("authors")
	cursor, err := authorsCollection.Find(ctx, bson.M{"_id": bson.In(authorIDs)})
	if err != nil {
		log.Fatal("Error finding authors:", err)
	}

	var authors []Author
	if err = cursor.All(ctx, &authors); err != nil {
		log.Fatal("Error decoding authors:", err)
	}

	authorMap := make(map[primitive.ObjectID]*Author)
	for _, author := range authors {
		authorMap[author.ID] = &author
	}

	for i, book := range books {
		book.author = authorMap[book.AuthorID]
	}
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")

	// Book and author setup (same as before)
	// ...

	eagerLoadingStart := time.Now()
	var booksWithAuthors []BookWithAuthor
	if err = fetchBooksWithAuthors(context.TODO(), db, &booksWithAuthors); err != nil {
		log.Fatal(err)
	}
	eagerLoadingEnd := time.Now()
	eagerLoadingDuration := eagerLoadingEnd.Sub(eagerLoadingStart).Milliseconds()

	fmt.Println("Eager loading performance:")