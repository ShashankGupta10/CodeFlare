package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Input struct {
    Url string `json:"url"`
}


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDeploy)
	mux.HandleFunc("POST /clone_github", handleURL)
	log.Println("Serving at port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleDeploy(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello Golang")
}

func handleURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusInternalServerError)
        return
    }
    var abc Input
    err = json.Unmarshal(body, &abc)
    if err != nil {
        fmt.Println("Error", err)
    }
    github_link := abc.Url
    return_var := handleClone(github_link)
    fmt.Println(return_var)
}

func handleClone(github_link string) string {
    return fmt.Sprintf("Hello %s", github_link)
}
