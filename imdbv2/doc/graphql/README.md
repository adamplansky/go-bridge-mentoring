https://dgraph.io/docs/graphql/mutations/delete/

```bigquery
type Movie {
title:       String!
release_date: DateTime!
duration: Int!
score:       Int!
description: String
artists: [Artist] @hasInverse(field: movies)
}

type Artist {
first_name:  String!
last_name:   String!
birth_date: DateTime!,
movies: [Movie] @hasInverse(field: artists)
}
```



mutation CreateMovie($movieInput: [AddMovieInput!]!) {
addMovie(input: $movieInput) {
movie {
title
duration
score
release_date
artists {
first_name
last_name
birth_date
}
}
}
}




{
"movieInput": [
{
"title": "Avengers: Endgame",
"duration": 182,
"score": 3,
"release_date": "2019-04-01T17:30:15+05:30",
"artists": [
{
}

type Artist {
first_name:  String!
last_name:   String!
birth_date: DateTime!,
movies: [Movie] @hasInverse(field: artists
"first_name": "Robert",
"last_name": "Downey Jr.",
"birth_date": "1965-04-04T17:30:15+05:30"
},
{
"first_name": "Scarlett",
"last_name": "Johansson",
"birth_date": "1984-11-22T17:30:15+05:30"
}

      ]
    },
    {
      "title": "22.11.1984",
      "duration": 124,
      "score": 7,
      "release_date": "2019-02-27T17:30:15+05:30",
      "artists": [
        {
          "first_name": "Scarlett",
          "last_name": "Johansson",
          "birth_date": "1984-11-22T17:30:15+05:30"
        }
        
      ]
    }
]
}

query MyQuery {
getMovie(id: "0x25700026b") {
id
duration
title
description
}
}






mutation DeleteMovie($filter: MovieFilter!){
deleteMovie(filter: $filter){
movie {
title
}
}
}


{ "filter":
{ "title": { "alloftext": "LOTR1" } }
}



mutation XXX($filter: ArtistFilter!){
deleteArtist(filter: $filter){
artist {
first_name
}
}
}

{ "filter":
{ "last_name": { "eq": "XXX" } }
}





mutation updateMovie($patch: UpdateMovieInput!) {
updateMovie(input: $patch) {
movie {
title
description
}
}
}


{
"patch": {
"filter": {
"title": {
"anyoftext": "22.11.1984"
}
},
"set": {
"title": "Captain Marvel"
}
}
}

-----------------
{
movie(func: has(Movie.title)) {
Movie.title
Movie.duration
Movie.release_date
Movie.artists {
Artist.birth_date
Artist.first_name
Artist.last_name
}
}
}

