package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// CommitData stores the commit information
type CommitData struct {
	Commits          []*github.RepositoryCommit `yaml:"commits"`
	Completed        []string                   `yaml:"completed"`
	LastRunTimestamp string                     `yaml:"last_run_timestamp"`
}

func (c *CommitData) IsCompleted(sha string) bool {
	for _, completed := range c.Completed {
		if completed == sha {
			return true
		}
	}
	return false
}

func (c *CommitData) AllCompleted() bool {
	for _, commit := range c.Commits {
		if !c.IsCompleted(commit.GetSHA()) {
			return false
		}
	}
	return true
}

func (c *CommitData) FormatDate(a *github.CommitAuthor) string {
	return a.GetDate().Format("01/02/2006 15:04")
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

	// Load existing commits from local storage
	existingCommits, err := LoadCommits(localFilename)
	if err == nil {
		fmt.Println("Loading commits from local storage")
		commitData = existingCommits
	}

	// Determine the starting point for fetching commits
	var since time.Time
	if len(commitData.Commits) == 0 {
		// If no commits loaded or if file doesn't exist, fetch the latest 30 commits
		fmt.Println("Fetching latest 30 commits from GitHub API")
		since = time.Now().AddDate(0, 0, -15) // fetch commits from the last 15 days
	} else {
		lastRunTime, err := time.Parse(time.RFC3339, commitData.LastRunTimestamp)
		if err != nil {
			return nil, fmt.Errorf("error parsing LastRunTimestamp: %w", err)
		}
		since = lastRunTime
	}

	// Fetch commits from GitHub API
	fmt.Println("Fetching commits from GitHub API")
	opts := &github.CommitsListOptions{Since: since}
	githubCommits, _, err := client.Repositories.ListCommits(ctx, config.RepoOwner, config.RepoName, opts)
	if err != nil {
		return nil, err
	}

	// Process each commit to fetch the diff (code changes)
	for _, githubCommit := range githubCommits {
		commit, _, err := client.Repositories.GetCommit(ctx, config.RepoOwner, config.RepoName, githubCommit.GetSHA(), nil)
		fmt.Println("GH Commit: ", commit)
		if err != nil {
			return nil, err
		}

		commitData.Commits = append(commitData.Commits, commit)
	}

	// Update the LastRunTimestamp or LastCommitHash in the config
	commitData.LastRunTimestamp = time.Now().Format(time.RFC3339)

	// Save fetched commits and diffs locally
	err = SaveCommitData(localFilename, commitData)
	if err != nil {
		fmt.Printf("Error saving commits and diffs: %v\n", err)
		return nil, err
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
