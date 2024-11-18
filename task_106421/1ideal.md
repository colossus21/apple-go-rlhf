Here's an example that demonstrates eager loading of related collections in MongoDB using the official Go driver. We'll implement a Books and Authors relationship where each Book has an associated Author. The example shows how to fetch Books along with their Author details in a single query using MongoDB's aggregation pipeline.

**Note:** In MongoDB, eager loading is implemented using the `$lookup` aggregation stage, which performs a left outer join between collections. This is different from traditional SQL ORMs where eager loading is built into the ORM itself.

```go
[Previous code remains the same]
```

**Key Implementation Details:**

1. **Data Models:**
    - `Author`: Base model for author information
    - `Book`: Base model with just the author reference (AuthorID)
    - `BookWithAuthor`: Special model for eager loading that includes the full Author details

2. **MongoDB Aggregation:**
    - Uses `$lookup` to join Books with Authors collection
    - `$unwind` flattens the author array (since `$lookup` returns an array)

3. **Key Operations:**
    - Connects to MongoDB with proper error handling
    - Inserts sample authors and books
    - Performs eager loading using aggregation pipeline
    - Retrieves and displays the combined data

This implementation demonstrates a clean way to handle related data in MongoDB while maintaining proper separation between base and eager-loaded models.

Would you like to see how to handle specific aspects like error handling or different relationship types?