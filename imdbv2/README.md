https://dgraph.io/docs/

https://pkg.go.dev/github.com/dgraph-io/dgo/v210

sql to graphql
https://dgraph.io/learn/courses/datamodel/sql-to-dgraph/develop/read-data/filter-results/


```shell
docker run --rm -it -p "8080:8080" -p "9080:9080" -p "8000:8000" -v ~/dgraph:/dgraph "dgraph/standalone:v21.03.0"
open http://localhost:8000/
```

# playground
https://play.dgraph.io/

```shell
curl -X POST http://localhost:8888/admin/schema --data-binary '@schema.graphql'
```

client:
snap install altair



http://localhost:8888/graphql

https://dgraph.io/docs/graphql/quick-start/

Mutation
```graphql
mutation {
  addProduct(input: [
    { name: "GraphQL on Dgraph"},
    { name: "Dgraph: The GraphQL Database"}
  ]) {
    product {
      productID
      name
    }
  }
  addCustomer(input: [{ username: "Michael"}]) {
    customer {
      username
    }
  }
}
```
```graphql
mutation {
    addReview(input: [{
        by: {username: "Michael"},
        about: { productID: "0x2"},
        comment: "Fantastic, easy to install, worked great.  Best GraphQL server available",
        rating: 10}])
    {
        review {
            comment
            rating
            by { username }
            about { name }
        }
    }
}
```

Query
```graphql
query {
  queryReview(filter: { comment: {alloftext: "easy to install"}}) {
    comment
    by {
      username
    }
    about {
      name
    }
  }
}
```

```graphql
query {
    queryCustomer(filter: { username: { regexp: "/Mich.*/" } }) {
        username
        reviews(order: { asc: rating }, first: 5) {
            comment
            rating
            about {
                name
            }
        }
    }
}
```

# GraphQL mutation
```graphql
mutation CreateMovie($movieInput: [AddMovieInput!]!) {
  addMovie(input: $movieInput) {
    movie {
      title
    }
  }
}


{
  "movieInput": [
    {
      "title": "LOTR2"
    },
    {
      "title": "LOTR3"
    }
  ]
}
```


1) create schema in cloud
2) add data
3) move query to code
4) conert dql to grapql