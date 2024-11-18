//go:build 2a
// +build 2a

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Author struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

type Book struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Title         string             `bson:"title"`
	AuthorID      primitive.ObjectID `bson:"author_id"`
	author        *Author            `bson:"author"`
	authorCache   atomic.Value
	forEachResult sync.WaitGroup
}

func (b *Book) GetCachedAuthor() *Author {
	cachedAuthor := b.authorCache.Load()
	return cachedAuthor.(*Author)
}

func (b *Book) GetAuthor(db *mongo.Database) {
	if author := b.GetCachedAuthor(); author != nil {
		log.Printf("Book %s author using Cache: %s\n", b.Title, author.Name)
		return
	}
	b.forEachResult.Add(1)
	go func() {
		defer b.forEachResult.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		col := db.Collection("authors")
		result := col.FindOne(ctx, bson.M{"_id": b.AuthorID})
		if result.Err() != nil {
			log.Println(err)
			return
		}
		var author Author
		if err := result.Decode(&author); err != nil {
			log.Println(err)
			return
		}
		b.authorCache.Store(&author)
		log.Printf("Book %s author using DB: %s\n", b.Title, author.Name)
	}()
}
func (b *Book) PreloadAuthor(db *mongo.Database, books []*Book) {
	authorIds := make([]primitive.ObjectID, len(books))
	for i, v := range books {
		authorIds[i] = v.AuthorID
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	col := db.Collection("authors")
	cursor, err := col.FindOneAndUpdate(ctx, bson.M{}, bson.M{"$set": bson.M{"works": authorIds}}, options.FindOneAndUpdate().SetReturnDocument(mongo.After))
	if err != nil {
		log.Println(err)
		return
	}
	if err := cursor.Decode(&authorsForBooks); err != nil {
		log.Println(err)
		return
	}
}

var authorsForBooks struct {
	ID    primitive.ObjectID   `bson:"_id"`
	Name  string               `bson:"name"`
	Works []primitive.ObjectID `bson:"works"`
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")

	cleanupCollections(db)
	prepareData(db)

	// Eager Loading
	fmt.Println("------------------- Eager Loading -------------------")
	start := time.Now()
	booksWithAuthors, err := eagerLoadBooks(db)
	if err != nil {
		log.Fatal(err)
	}
	eagerLoadDuration := time.Since(start)
}
