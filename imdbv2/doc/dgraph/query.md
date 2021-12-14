```dql
{
 me(func: has(starring)) {
   name
  }
}
```


# Get all movies released after “1980”
```dql  
{
  me(func: allofterms(name, "Star Wars"), orderasc: release_date) @filter(ge(release_date, "1980")) {
    name
    release_date
    revenue
    running_time
    director {
     name
    }
    starring (orderasc: name) {
     name
    }
  }
}
```
