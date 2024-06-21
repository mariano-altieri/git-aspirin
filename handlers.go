package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

func Serve() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", reportHandler)
	http.HandleFunc("/resolve", resolveHandler)
	fmt.Println("Serving report at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func resolveHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CommitHash string `json:"commit"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if body.CommitHash == "" {
		http.Error(w, "commit parameter is required", http.StatusBadRequest)
		return
	}

	commitData, err := LoadCommits("commits.yaml")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading commits: %v", err), http.StatusInternalServerError)
		return
	}

	found := false
	for _, completed := range commitData.Completed {
		if completed == body.CommitHash {
			found = true
			break
		}
	}

	if !found {
		commitData.Completed = append(commitData.Completed, body.CommitHash)
	} else {
		// remove the commit from the completed list
		var newCompleted []string
		for _, completed := range commitData.Completed {
			if completed != body.CommitHash {
				newCompleted = append(newCompleted, completed)
			}
		}
		commitData.Completed = newCompleted
	}

	err = SaveCommitData("commits.yaml", commitData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving commits: %v", err), http.StatusInternalServerError)
		return
	}
}

func reportHandler(w http.ResponseWriter, _ *http.Request) {
	commitData, err := LoadCommits("commits.yaml")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading commits: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("template.html"))

	err = tmpl.Execute(w, commitData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
		return
	}
}
