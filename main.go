package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/jedib0t/go-pretty/table"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func main() {
	dirPath := os.Args[1]

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	var rows []table.Row

	for _, f := range files {
		if f.IsDir() {
			absPath := path.Join(dirPath, f.Name())
			branch, tag, err1 := GetCurrentBranchAndTagFromPath(absPath)
			remoteNames, err2 := GetRemotes(absPath)
			head, err3 := GetCurrentCommitFromPath(absPath)
			if err1 == nil && err2 == nil && err3 == nil {
				if len(remoteNames) == 0 {
					rows = append(rows, table.Row{f.Name(), head[:7], branch, tag})
					continue
				} else {
					rows = append(rows, table.Row{f.Name(), head[:7], branch, tag, remoteNames[0]})
				}
				for i, remoteName := range remoteNames {
					if i == 0 {
						continue
					}
					rows = append(rows, table.Row{"", "", "", remoteName})
				}
			}
		}
	}

	// Print the branch names, tags and remotes in a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Repository", "Head", "Branch", "Tag", "Remotes"})
	t.AppendRows(rows)
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

func GetRemotes(path string) ([]string, error) {
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
			latestTagName = tagRef.Name().String()
		}

		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = tagRef.Name().String()
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return latestTagName, nil
}
