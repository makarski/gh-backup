package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/makarski/gh-backup/config"
	"github.com/makarski/gh-backup/github"
)

var (
	wg  sync.WaitGroup
	sem = make(chan int, 15)

	infoLog = log.New(os.Stdout, "[gh-backup] INFO: ", log.LUTC)
	errLog  = log.New(os.Stderr, "[gh-backup] ERR: ", log.LUTC)
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		errLog.Panic(err)
	}

	ghClient := github.NewClient(
		cfg.AccessToken,
		&github.ClientOptions{
			Languages: cfg.Languages,
		},
	)

	infoLog.Println("listing repos for:", cfg.Organisation, "languages:", cfg.Languages)

	repos, err := ghClient.ListOrgRepos(cfg.Organisation, cfg.OnlyPrivate)
	if err != nil {
		errLog.Panic(err)
	}

	infoLog.Println("found repos:", len(repos))

	if err := createBackupDir(cfg.BackupDir); err != nil {
		errLog.Panic(err)
	}

	for _, repo := range repos {
		sem <- 1
		wg.Add(1)
		go processRepo(repo, cfg.BackupDir)
	}

	wg.Wait()
}

func processRepo(ghRepo github.Repo, backupDir string) {
	defer func() {
		wg.Done()
		<-sem
	}()

	projectDir := filepath.Join(backupDir, ghRepo.Name)

	if dirExists(projectDir) {
		infoLog.Println("updating:", ghRepo.FullName)
		if err := updateRepo(projectDir, ghRepo.SshURL); err != nil {
			errLog.Println("update:", err)
		}
		return
	}

	infoLog.Println("cloning:", ghRepo.FullName)
	if err := cloneRepo(backupDir, ghRepo.SshURL); err != nil {
		errLog.Println("clone:", err)
	}
}

func createBackupDir(backupDir string) error {
	return os.MkdirAll(backupDir, os.FileMode(0755))
}

func dirExists(targetDir string) bool {
	_, err := os.Stat(targetDir)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func cloneRepo(targetDir, sshURL string) error {
	cmd := exec.Command("git", "-C", targetDir, "clone", sshURL)
	return cmd.Run()
}

func updateRepo(targetDir, sshURL string) error {
	cmd := exec.Command("git", "-C", targetDir, "pull", "--all")
	return cmd.Run()
}
