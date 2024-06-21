package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// CommitData stores the commit information
type CommitData struct {
	Commits   []*github.RepositoryCommit `yaml:"commits"`
	Completed []string                   `yaml:"completed"`
}

func (c *CommitData) IsCompleted(sha string) bool {
	for _, completed := range c.Completed {
		if completed == sha {
			return true
		}
	}
	return false
}

func AuthenticateGitHub(config *Config) (c context.Context, cl *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.GitHubToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return ctx, client
}

// fetchCommits fetches commits and their diffs either from local storage or GitHub API
func FetchCommits(ctx context.Context, client *github.Client, config *Config) (*CommitData, error) {
	// Check if commits are stored locally
	localFilename := "commits.yaml"
	commitData := &CommitData{}

	commits, err := LoadCommits(localFilename)

	if err == nil {
		fmt.Println("Loading commits from local storage")
		return commits, nil
	}

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
