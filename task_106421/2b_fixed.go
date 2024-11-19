//go:build 2ideal
// +build 2ideal

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

var (
	authorCache = make(map[primitive.ObjectID]*Author)
	mu          sync.RWMutex
	db          *mongo.Database
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

func (b *Book) GetCachedAuthor() (*Author, error) {
	mu.RLock()
	defer mu.RUnlock()
	if author, exists := authorCache[b.AuthorID]; exists {
		return author, nil
	}
	return nil, fmt.Errorf("author not found in cache")
}

func (b *Book) GetAuthor(ctx context.Context) (*Author, error) {
	if author, err := b.GetCachedAuthor(); err == nil {
		return author, nil
	}

	var author Author
	err := db.Collection("authors").FindOne(ctx, bson.M{"_id": b.AuthorID}).Decode(&author)
	if err != nil {
		return nil, fmt.Errorf("error fetching author: %v", err)
	}

	mu.Lock()
	authorCache[b.AuthorID] = &author
	mu.Unlock()

	return &author, nil
}

func PreloadAuthors(ctx context.Context, books []*Book) error {
	if len(books) == 0 {
		return nil
	}

	ids := make([]primitive.ObjectID, len(books))
	for i, book := range books {
		ids[i] = book.AuthorID
	}

	cursor, err := db.Collection("authors").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return fmt.Errorf("error fetching authors: %v", err)
	}
	defer cursor.Close(ctx)

	mu.Lock()
	defer mu.Unlock()

	for cursor.Next(ctx) {
		var author Author
		if err := cursor.Decode(&author); err != nil {
			return fmt.Errorf("error decoding author: %v", err)
		}
		authorCache[author.ID] = &author
	}

	return cursor.Err()
}

func main() {
	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db = client.Database("mydatabase")
	if err := db.Drop(ctx); err != nil {
		log.Fatal(err)
	}

	authorCount := 100
	bookCount := 1000

	authors := createAuthors(ctx, authorCount)
	createBooks(ctx, authors, bookCount)

	for i := 0; i < 3; i++ {
		fmt.Printf("\nIteration %d:\n", i+1)
		runLoadingTests(ctx, authorCount, bookCount)
	}
}

func createAuthors(ctx context.Context, count int) []Author {
	authors := make([]Author, count)
	authorsI := make([]interface{}, count)

	for i := 0; i < count; i++ {
		authors[i] = Author{ID: primitive.NewObjectID(), Name: fmt.Sprintf("Author %d", i+1)}
		authorsI[i] = authors[i]
	}

	_, err := db.Collection("authors").InsertMany(ctx, authorsI)
	if err != nil {
		log.Fatal(err)
	}

	return authors
}

func createBooks(ctx context.Context, authors []Author, count int) {
	booksI := make([]interface{}, count)
	for i := 0; i < count; i++ {
		authorIndex := rand.Intn(len(authors))
		booksI[i] = Book{
			ID:       primitive.NewObjectID(),
			Title:    fmt.Sprintf("Book %d", i+1),
			AuthorID: authors[authorIndex].ID,
		}
	}

	_, err := db.Collection("books").InsertMany(ctx, booksI)
	if err != nil {
		log.Fatal(err)
	}
}

func runLoadingTests(ctx context.Context, authorCount, bookCount int) {
	mu.Lock()
	authorCache = make(map[primitive.ObjectID]*Author)
	mu.Unlock()

	fmt.Println("Eager Loading Test:")
	start := time.Now()
	var booksWithAuthors []BookWithAuthor
	pipeline := mongo.Pipeline{
		bson.D{{"$lookup", bson.D{
			{"from", "authors"},
			{"localField", "author_id"},
			{"foreignField", "_id"},
			{"as", "author"},
		}}},
		bson.D{{"$unwind", "$author"}},
	}

	cursor, err := db.Collection("books").Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(ctx, &booksWithAuthors); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Eager loaded %d books and %d authors in %v\n", bookCount, authorCount, time.Since(start))

	fmt.Println("\nLazy Loading Test:")
	start = time.Now()
	var books []*Book
	cursor, err = db.Collection("books").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(ctx, &books); err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		_, err := book.GetAuthor(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Lazy loaded %d books and %d authors in %v\n", bookCount, authorCount, time.Since(start))

	fmt.Println("\nBatch Lazy Loading Test:")
	start = time.Now()
	authorCache = make(map[primitive.ObjectID]*Author)
	cursor, err = db.Collection("books").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(ctx, &books); err != nil {
		log.Fatal(err)
	}

	if err := PreloadAuthors(ctx, books); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Batch lazy loaded %d books and %d authors in %v\n", bookCount, authorCount, time.Since(start))
}
