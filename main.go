package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func main() {
	var dirPath string
	if len(os.Args) == 1 {
		dirPath = "./"
	} else if len(os.Args) == 2 {
		dirPath = os.Args[1]
	} else {
		Info("Usage %s [path to the parent folder of git repos]", os.Args[0])
		return
	}

	var rows []table.Row

	err := filepath.Walk(dirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Skips .git folder
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}

			var level int
			level, err = GetChildLevel(dirPath, path)
			if err != nil {
				panic(err)
			}

			if info.IsDir() && level <= 2 {
				branch, tag, err1 := GetCurrentBranchAndTagFromPath(path)
				remoteNames, err2 := GetRemotesFromPath(path)
				head, err3 := GetCurrentCommitFromPath(path)
				latest_tag, err4 := GetLatestTagFromPath(path)
				if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
					if len(remoteNames) == 0 {
						rows = append(rows, table.Row{path, head[:7], branch, tag, latest_tag, ""})
					} else {
						rows = append(rows, table.Row{path, head[:7], branch, tag, latest_tag, strings.Join(remoteNames, "\n")})
					}
				}
			}
			return nil
		})
	CheckIfError(err)

	// Print the branch names, tags and remotes in a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Repository", "Head", "Branch", "Tag", "Latest Tag", "Remotes"})
	t.AppendRows(rows)
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}

func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func CountLevel(s []byte) int {
	count := int(0)
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			count++
		}
	}
	return count
}

func GetChildLevel(basepath, targpath string) (int, error) {
	rel, err := filepath.Rel(basepath, targpath)
	if err != nil {
		return 0, err
	}
	return CountLevel([]byte(rel)) + 1, nil
}

func GetRemotesFromPath(path string) ([]string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	remotes, err := r.Remotes()
	remoteNames := make([]string, len(remotes))
	for i, remote := range remotes {
		remoteNames[i] = remote.String()
		// Info(remote.String())
	}
	return remoteNames, nil
}

func GetCurrentBranchAndTagFromPath(path string) (string, string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", "", err
	}
	return GetCurrentBranchAndTag(r)
}

func GetCurrentBranchAndTag(repository *git.Repository) (string, string, error) {
	branchRefs, err := repository.Branches()
	if err != nil {
		return "", "", err
	}

	headRef, err := repository.Head()
	if err != nil {
		return "", "", err
	}

	var currentBranchName string
	branchRefs.ForEach(func(branchRef *plumbing.Reference) error {
		if branchRef.Hash() == headRef.Hash() {
			currentBranchName = branchRef.Name().Short()
			return nil
		}
		return nil
	})

	var currentTagName string
	tagRefs, err := repository.Tags()
	tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		if tagRef.Hash() == headRef.Hash() {
			currentTagName = tagRef.Name().Short()
			return nil
		}
		return nil
	})

	return currentBranchName, currentTagName, nil
}

func GetCurrentCommitFromPath(path string) (string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}
	return GetCurrentCommit(r)
}

func GetCurrentCommit(repository *git.Repository) (string, error) {
	headRef, err := repository.Head()
	if err != nil {
		return "", err
	}
	headSha := headRef.Hash().String()

	return headSha, nil
}

func GetLatestTagFromPath(path string) (string, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}
	return GetLatestTagFromRepository(r)
}

func GetLatestTagFromRepository(repository *git.Repository) (string, error) {
	tagRefs, err := repository.Tags()
	if err != nil {
		return "", err
	}

	var latestTagCommit *object.Commit
	var latestTagName string
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		revision := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repository.ResolveRevision(revision)
		if err != nil {
			return err
		}

		commit, err := repository.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}

		if latestTagCommit == nil {
			latestTagCommit = commit
			latestTagName = tagRef.Name().Short()
		}

		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = tagRef.Name().Short()
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return latestTagName, nil
}
