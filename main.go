package main

import (
	// "context"
	// "bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-github/v62/github"
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
	Commits []*github.RepositoryCommit
}

// fetchCommits fetches commits and their diffs either from local storage or GitHub API
func fetchCommits(ctx context.Context, client *github.Client, config *Config) (*CommitData, error) {
	// Check if commits are stored locally
	localFilename := "commits.yaml"
	commitData := &CommitData{}

	commits, err := LoadCommits(localFilename)
	fmt.Println(">>>>>", commits)
	if err == nil {
		fmt.Println("Loading commits from local storage")
		commitData.Commits = commits
		return commitData, nil
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
	data, err := yaml.Marshal(commitData.Commits)
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
func LoadCommits(filename string) ([]*github.RepositoryCommit, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println("Error reading commits file")
		return nil, err
	}
	var commits []*github.RepositoryCommit
	err = yaml.Unmarshal(data, &commits)
	fmt.Println(">>>>> commits", commits)
	fmt.Println(">>>>> err", err)

	if err != nil {
		fmt.Println("Error unmarshalling commits")
		return nil, err
	}
	return commits, nil
}

// getLocalFiles recursively gets all files in the given directory, excluding specified folders
func getLocalFiles(root string, excludeFolders []string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Check if the directory should be excluded
			for _, folder := range excludeFolders {
				if info.Name() == folder {
					return filepath.SkipDir
				}
			}
		} else {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// getChangedFilesFromCommit fetches the files changed in a specific commit
func getChangedFilesFromCommit(ctx context.Context, client *github.Client, owner, repo, sha string) ([]string, error) {
	commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, file := range commit.Files {
		files = append(files, file.GetFilename())
	}
	return files, nil
}

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

func main() {
	// Load the configuration
	// config, err := LoadConfig("config.yaml")
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Get the local files
	localFiles, err := getLocalFiles(config.LocalRepoPath, config.ExcludeFolders)
	if err != nil {
		log.Fatalf("Error getting local files: %v", err)
	}

	// Set up GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Fetch the latest commits
	commitData, err := fetchCommits(ctx, client, config)
	if err != nil {
		log.Fatalf("Error fetching commits: %v", err)
	}

	// // Map to store the latest changed files
	changedFiles := make(map[string]bool)

	// Print the commit hash and description
	// for _, commit := range commits {
	// 	sha := commit.GetSHA()
	// 	message := commit.Commit.GetMessage()
	// 	author := commit.Commit.Author.GetName()
	// 	fmt.Printf("Commit: %s\nDescription: %s\nAuthor: %s\n\n", sha, message, author)
	// }

	// Get changed files from the latest commits
	for _, commit := range commitData.Commits {
		files, err := getChangedFilesFromCommit(ctx, client, config.RepoOwner, config.RepoName, commit.GetSHA())
		if err != nil {
			log.Fatalf("Error getting changed files from commit: %v", err)
		}
		for _, file := range files {
			changedFiles[file] = true
		}
	}

	// // Compare changed files with local files
	fmt.Println("Changed files:")
	for file := range changedFiles {
		fmt.Printf("Remote file: %s\n", file)
	}

	fmt.Println("Local files:")
	for _, file := range localFiles {
		fmt.Printf("Local file: %s\n", file)
	}
}
