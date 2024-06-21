# Git Aspirin

CLI tool to assist in the migration of legacy code to a new structure.


## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Installation](#installation)
- [Roadmap](#roadmap)

## Overview

Git Aspirin is a command-line tool designed to assist in the migration of legacy code to a new structure. It monitors GitHub repositories, compares recent commits to a local target repository, and highlights new code additions for manual porting. This tool aims to streamline the modernization process by focusing on code changes and facilitating manual updates.

## Tech Stack

- Go (Golang): Utilized for developing the CLI tool due to its efficiency and concurrency features.
- GitHub API (go-github): Integrated to fetch commits and code changes from GitHub repositories.
- HTML/CSS/JavaScript: Used for generating and styling HTML reports to visualize commit differences.
- YAML: Configured for storing tool configurations and commit data locally.
- Cobra: Employed for creating CLI commands and managing tool functionality.

## Installation

### Development Setup

- Ensure you have Go installed. If not, follow the official [installation guide](https://go.dev/doc/install).
- Clone the Repository
```bash
git clone https://github.com/mariano-altieri/git-aspirin.git .
cd git-aspirin
```
- Rename `config.yaml.example` to `config.yaml` and update it with your GitHub credentials and repository details. Generate a new Personal Access Token [here](https://github.com/settings/developers).
- If you aim to use `git-aspirin` with a private repo you don't own, send the Personal Access Token to the Repository Owner to add it.
- Build and Run
```bash
go build -o git-aspirin *.go
```

### Binary Installation

- Download the latest release from the [Releases]()

- Rename `config.yaml.example` to `config.yaml` and update it with your GitHub credentials and repository details. Generate a new Personal Access Token [here](https://github.com/settings/developers).
- If you aim to use `git-aspirin` with a private repo you don't own, send the Personal Access Token to the Repository Owner to add it.
- Run the binary
```bash
./git-aspirin run
```

## Roadmap
- Improve code
- Add tests
- Add more features
- Add more documentation
