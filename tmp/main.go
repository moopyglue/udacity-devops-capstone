
package main

import (
	"fmt"
    "os"
)

var work_queue = make(chan string,100)

type message struct {
    

func main() {

    done := make(chan bool,1)

    go func() {
      s := ""
      for {
        s = <-work_queue
        if s == "" { break }
        fmt.Println("[",s,"]")
      }
      done <-true
    }()

    for _, pair := range os.Environ() {
        work_queue <- pair
    }
    work_queue <- ""
    <-done

    os.Exit(0)
}

