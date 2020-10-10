package main

import (
    "fmt"
	"time"
)

type Post struct {
    Id     int
    Content string
    Author  string
}

var PostById map[int]*Post
var PostsByAuthor map[string][]*Post

func store(post Post) {
    PostById[post.Id] = &post
    PostsByAuthor[post.Author] = append(PostsByAuthor[post.Author], &post)
}

func main() {


    PostById = make(map[int]*Post)
    PostsByAuthor = make(map[string][]*Post)

    post1 := Post{Id: 1, Content: "Hello World!", Author: "Sau Sheong"}
    post2 := Post{Id: 2, Content: "Bonjour Monde!", Author: "Pierre"}
    post3 := Post{Id: 3, Content: "Hola Mundo!", Author: "Pedro"}
    post4 := Post{Id: 4, Content: "Greetings Earthlings!", Author: "Sau Sheong"}

    store(post1)
    store(post2)
    store(post3)
    store(post4)


    fmt.Println(PostById[1])
    fmt.Println("")
    fmt.Println(PostById[2])
    fmt.Println("")

    for _, post := range PostsByAuthor["Sau Sheong"] {
        fmt.Println(post)
    }
    fmt.Println("")
    for _, post := range PostsByAuthor["Pedro"] {
        fmt.Println(post)
    }

	start := time.Now()
	end := time.Now()
	diff := end.Sub(start)
	fmt.Printf("start = %d\n", start.UnixNano())
	fmt.Printf("end = %d\n", end.UnixNano())
	fmt.Printf("difference = %d\n", diff)
}
