//go:build 2b
// +build 2b

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var authorCache = make(map[primitive.ObjectID]Author)
var mu sync.RWMutex

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

func (b *Book) GetCachedAuthor() *Author {
	mu.RLock()
	defer mu.RUnlock()
	return &authorCache[b.AuthorID]
}

func (b *Book) GetAuthor() {
	if author := b.GetCachedAuthor(); author != nil {
		return
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")
	authorsCollection := db.Collection("authors")

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err = authorsCollection.FindOne(ctx, bson.M{"_id": b.AuthorID}).Decode(&authorCache[b.AuthorID])
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Book) PreloadAuthor(books []*Book) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")
	authorsCollection := db.Collection("authors")

	ids := make([]primitive.ObjectID, len(books))
	for i, book := range books {
		ids[i] = book.AuthorID
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	cur, err := authorsCollection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(context.TODO())

	for cur.Next(ctx) {
		var author Author
		if err = cur.Decode(&author); err != nil {
			log.Fatal(err)
		}
		mu.Lock()
		authorCache[author.ID] = author
		mu.Unlock()
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mydatabase")
	booksCollection := db.Collection("books")
	authorsCollection := db.Collection("authors")

	authorCount := 100
	bookCount := 1000

	// Create authors and books
	authors := createAuthors(authorCount)
	createBooks(booksCollection, authors, bookCount)

	fmt.Println("Eager Loading Test:")
	eagerLoadingTest(booksCollection, authorCount, bookCount)

	fmt.Println("\nLazy Loading Test:")
	lazyLoadingTest(booksCollection, authorCount, bookCount)

	fmt.Println("\nBatch Lazy Loading Test:")
	batchLazyLoadingTest(booksCollection, authorCount, bookCount)
}

func createAuthors(count int) []Author {
	authors := make([]Author, count)
	for i := 0; i < count; i++ {
		authors[i] = Author{ID: primitive.NewObjectID(), Name: fmt.Sprintf("Author %d", i+1)}
	}
	return authors
}

func createBooks(booksCollection *mongo.Collection, authors []Author, count int) {
	books := make([]Book, count)
	for i := 0; i < count; i++ {
		authorIndex := rand.Intn(len(authors))
		books[i] = Book{ID: primitive.NewObjectID(), Title: fmt.Sprintf("Book %d", i+1), AuthorID: authors[authorIndex].ID}
	}

	_, err := booksCollection.InsertMany(context.TODO(), books)
	if err != nil {
		log.Fatal("Error inserting books:", err)
	}
}

func eagerLoadingTest(booksCollection *mongo.Collection, authorCount, bookCount int) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	start := time.Now()
	var booksWithAuthors []BookWithAuthor
	err := booksCollection.Aggregate(ctx, []bson.M{
		{"$lookup": bson.M{
			"from":         "authors",
			"localField":   "author_id",
			"foreignField": "_id",
			"as":           "author",
		}},
		{"$unwind": "$author"},
	}).All(ctx, &booksWithAuthors)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	fmt.Printf("Eager loaded %d books and %d authors in %v\n", bookCount, authorCount, elapsed)
}

func lazyLoadingTest(booksCollection *mongo.Collection, authorCount, bookCount int) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	start := time.Now()
	var books []Book
	err := booksCollection.Find(ctx, bson.M{}).All(ctx, &books)
	if err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		book.GetAuthor()
	}

	elapsed := time.Since(start)
	fmt.Printf("Lazy loaded %d books and %d authors in %v\n", bookCount, authorCount, elapsed)
}

func batchLazyLoadingTest(booksCollection *mongo.Collection, authorCount, bookCount int) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	start := time.Now()
	var books []Book
	err := booksCollection.Find(ctx, bson.M{}).All(ctx, &books)
	if err != nil {
		log.Fatal(err)
	}

	book.PreloadAuthor(books)

	elapsed := time.Since(start)
	fmt.Printf("Batch lazy loaded %d books and %d authors in %v\n", bookCount, authorCount, elapsed)
}
