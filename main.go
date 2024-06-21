package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
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

		ctx, client := AuthenticateGitHub(config)

		_, err = FetchCommits(ctx, client, config)
		if err != nil {
			log.Fatalf("Error fetching commits: %v", err)
		}

		Serve()
	},
}

func main() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
