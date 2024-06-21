package main

import (
	// "context"
	// "bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the GitHub client
type Config struct {
	GitHubToken    string   `yaml:"github_token"`
	RepoOwner      string   `yaml:"repo_owner"`
	RepoName       string   `yaml:"repo_name"`
	LocalRepoPath  string   `yaml:"local_repo_path"`
	ExcludeFolders []string `yaml:"exclude_folders"`
}

// CommitData stores the commit information
type CommitData struct {
	Commits   []*github.RepositoryCommit `yaml:"commits"`
	Completed []string                   `yaml:"completed"`
}

// type PageData struct {
// 	Commits   []*github.RepositoryCommit
// 	Completed []CommitCompleted
// }

func generateHTMLReport(commits *CommitData) error {
	tmpl := template.Must(template.ParseFiles("template.html"))
	// data := struct {
	// 	Commits []*github.RepositoryCommit
	// }{
	// 	Commits: commits.Commits,
	// }

	f, err := os.Create("output.html")
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, commits)
	if err != nil {
		return err
	}

	return nil
}

func serveReport() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "output.html")
	})
	http.HandleFunc("/resolve", resolveHandler)
	fmt.Println("Serving report at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// fetchCommits fetches commits and their diffs either from local storage or GitHub API
func fetchCommits(ctx context.Context, client *github.Client, config *Config) (*CommitData, error) {
	// Check if commits are stored locally
	localFilename := "commits.yaml"
	commitData := &CommitData{}

	commits, err := LoadCommits(localFilename)

	if err == nil {
		fmt.Println("Loading commits from local storage")
		return commits, nil
	}

	// return nil, nil

	// Fetch commits from GitHub API
	fmt.Println("Fetching commits from GitHub API")
	githubCommits, _, err := client.Repositories.ListCommits(ctx, config.RepoOwner, config.RepoName, nil)
	if err != nil {
		return nil, err
	}

	// Process each commit to fetch the diff (code changes)
	for _, githubCommit := range githubCommits {
		commit, _, err := client.Repositories.GetCommit(ctx, config.RepoOwner, config.RepoName, githubCommit.GetSHA(), nil)
		if err != nil {
			return nil, err
		}
		// fmt.Println(commit)
		commitData.Commits = append(commitData.Commits, commit)
	}

	// Save fetched commits and diffs locally
	err = SaveCommitData(localFilename, commitData)
	if err != nil {
		fmt.Printf("Error saving commits and diffs: %v\n", err)
		// Handle error saving commits and diffs (optional)
	}

	return commitData, nil
}

// SaveCommitData saves fetched commits and diffs to a local file
func SaveCommitData(filename string, commitData *CommitData) error {
	data, err := yaml.Marshal(commitData)
	if err != nil {
		log.Panic("Error marshalling commit data")
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Panic("Error writing commit data to file")
		return err
	}
	return nil
}

// SaveCommits saves fetched commits to a local file
// func SaveCommits(filename string, commits []*github.RepositoryCommit) error {
// 	data, err := yaml.Marshal(commits)
// 	if err != nil {
// 		return err
// 	}
// 	err = os.WriteFile(filename, data, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// LoadCommits loads commits from a local file
func LoadCommits(filename string) (*CommitData, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var commitData CommitData
	//var commits []*github.RepositoryCommit
	err = yaml.Unmarshal(data, &commitData)

	if err != nil {
		fmt.Println("Error unmarshalling commits")
		return nil, err
	}
	return &commitData, nil
}

// getLocalFiles recursively gets all files in the given directory, excluding specified folders
// func getLocalFiles(root string, excludeFolders []string) ([]string, error) {
// 	var files []string
// 	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if info.IsDir() {
// 			// Check if the directory should be excluded
// 			for _, folder := range excludeFolders {
// 				if info.Name() == folder {
// 					return filepath.SkipDir
// 				}
// 			}
// 		} else {
// 			files = append(files, path)
// 		}
// 		return nil
// 	})
// 	return files, err
// }

// getChangedFilesFromCommit fetches the files changed in a specific commit
// func getChangedFilesFromCommit(ctx context.Context, client *github.Client, owner, repo, sha string) ([]string, error) {
// 	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var files []string
// 	for _, file := range commit.Files {
// 		files = append(files, file.GetFilename())
// 	}
// 	return files, nil
// }

// fetchCommits fetches commits either from local storage or GitHub API
// func fetchCommits(ctx context.Context, client *github.Client, config *Config) ([]*github.RepositoryCommit, error) {
// 	// Check if commits are stored locally
// 	localFilename := "commits.yaml"
// 	commits, err := LoadCommits(localFilename)
// 	if err == nil {
// 		fmt.Println("Loading commits from local storage")
// 		return commits, nil
// 	}

// 	// Fetch commits from GitHub API
// 	fmt.Println("Fetching commits from GitHub API")
// 	commits, _, err = client.Repositories.ListCommits(ctx, config.RepoOwner, config.RepoName, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Save fetched commits locally
// 	err = SaveCommits(localFilename, commits)
// 	if err != nil {
// 		fmt.Printf("Error saving commits: %v\n", err)
// 		// Handle error saving commits (optional)
// 	}

// 	return commits, nil
// }

// LoadConfig reads the configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		log.Fatal("Could not read config.yaml file. Please make sure it exists.")
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(data, &config)

	if err != nil {
		log.Fatal("Could not parse config.yaml file.")
		return nil, err
	}

	return &config, nil
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

	// fmt.Fprintf(w, "Commit %s marked as completed\n", body.CommitHash)
}

var rootCmd = &cobra.Command{
	Use:   "git-aspirin",
	Short: "Git Aspirin CLI tool",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch commits and generate an HTML report",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := LoadConfig("config.yaml")
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}

		ctx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.GitHubToken})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		commits, err := fetchCommits(ctx, client, config)
		if err != nil {
			log.Fatalf("Error fetching commits: %v", err)
		}

		err = generateHTMLReport(commits)
		if err != nil {
			log.Fatalf("Error generating HTML report: %v", err)
		}

		// fmt.Printf("%d new commits have been found.\n", len(commits.Commits))
		// You may also want to count modified files

		serveReport()
	},
}

func main() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Load the configuration
	// config, err := LoadConfig("config.yaml")
	// if err != nil {
	// 	log.Fatalf("Error loading configuration: %v", err)
	// }

	// Get the local files
	// localFiles, err := getLocalFiles(config.LocalRepoPath, config.ExcludeFolders)
	// if err != nil {
	// 	log.Fatalf("Error getting local files: %v", err)
	// }

	// Set up GitHub client
	// ctx := context.Background()
	// ts := oauth2.StaticTokenSource(
	// 	&oauth2.Token{AccessToken: config.GitHubToken},
	// )
	// tc := oauth2.NewClient(ctx, ts)
	// client := github.NewClient(tc)

	// // Fetch the latest commits
	// commitData, err := fetchCommits(ctx, client, config)
	// if err != nil {
	// 	log.Fatalf("Error fetching commits: %v", err)
	// }

	// // Map to store the latest changed files
	// changedFiles := make(map[string]bool)

	// Create a new template and parse the HTML file
	// tmpl := template.Must(template.ParseFiles("template.html"))
	// // fmt.Println(">>>>>", len(commitData.Commits))

	// // Render the template with the fetched commits
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	data := PageData{
	// 		Commits: commitData.Commits,
	// 	}
	// 	err := tmpl.Execute(w, data)
	// 	if err != nil {
	// 		// http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// })

	// // Start the HTTP server
	// fmt.Println("Server started at http://localhost:8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))

	// Print the commit hash and description
	// for _, commit := range commits {
	// 	sha := commit.GetSHA()
	// 	message := commit.Commit.GetMessage()
	// 	author := commit.Commit.Author.GetName()
	// 	fmt.Printf("Commit: %s\nDescription: %s\nAuthor: %s\n\n", sha, message, author)
	// }

	// Get changed files from the latest commits
	// for _, commit := range commitData.Commits {
	// 	files, err := getChangedFilesFromCommit(ctx, client, config.RepoOwner, config.RepoName, commit.GetSHA())
	// 	if err != nil {
	// 		log.Fatalf("Error getting changed files from commit: %v", err)
	// 	}
	// 	for _, file := range files {
	// 		changedFiles[file] = true
	// 	}
	// }

	// // // Compare changed files with local files
	// fmt.Println("Changed files:")
	// for file := range changedFiles {
	// 	fmt.Printf("Remote file: %s\n", file)
	// }

	// fmt.Println("Local files:")
	// for _, file := range localFiles {
	// 	fmt.Printf("Local file: %s\n", file)
	// }
}
