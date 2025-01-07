package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jedib0t/go-pretty/table"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

func initFlagParser(path *string, level *int, color *bool, update *bool, showVersion *bool) {
	dir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	flag.StringVar(path, "path", dir, "The path to the parent directory of git repos.")
	flag.IntVar(level, "depth", 2, "The depth ggit should go searching.")
	flag.BoolVar(color, "color", true, "Whether the table should be rendered with color.")
	flag.BoolVar(update, "update", false, "Try go-github-selfupdate via GitHub")
	flag.BoolVar(showVersion, "version", false, "Show version")
}

const version = "0.2.3"
const slug = "maverobot/ggit"

func selfUpdate(slug string) error {
	selfupdate.EnableLog()

	previous := semver.MustParse(version)
	latest, err := selfupdate.UpdateSelf(previous, slug)
	if err != nil {
		return err
	}

	if previous.Equals(latest.Version) {
		info("\nCurrent binary is the latest version %s", version)
	} else {
		info("\nSuccessfully updated from version %s to version %s\n", version, latest.Version)
		info("Release note:\n%s", latest.ReleaseNotes)
	}
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: ggit [flags]")
	flag.PrintDefaults()
}

func main() {
	var dirPath string
	var depth int
	var color bool
	var update bool
	var showVersion bool

	initFlagParser(&dirPath, &depth, &color, &update, &showVersion)
	flag.Usage = usage
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(flag.Args()) != 0 {
		usage()
		os.Exit(0)
	}

	if update {
		if err := selfUpdate(slug); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	var rows []table.Row

	err := filepath.Walk(dirPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			//	Skips files and not permitted access
			if !info.IsDir() || !canRead(&info) {
				return nil
			}

			// hidden folders
			if isHidden(info.Name()) {
				return filepath.SkipDir
			}

			var level int
			level, err = getChildLevel(dirPath, path)
			if err != nil {
				panic(err)
			}

			if info.IsDir() && level <= depth {
				r, err := git.PlainOpen(path)
				if err != nil {
					return nil
				}

				branch, tag, err1 := getCurrentBranchAndTag(r)
				remoteNames, err2 := getRemotes(r)
				head, err3 := getCurrentCommit(r)
				latestTag, err4 := getLatestTag(r)
				relPath, err5 := filepath.Rel(dirPath, path)
				isClean, err6 := isClean(r)

				var cleannessPrefix string
				if isClean {
					cleannessPrefix = ""
				} else {
					cleannessPrefix = "*"
				}

				if err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil && err6 == nil {
					if len(remoteNames) == 0 {
						rows = append(rows, table.Row{relPath, cleannessPrefix + head[:7], branch, tag, latestTag, ""})
					} else {
						rows = append(rows, table.Row{relPath, cleannessPrefix + head[:7], branch, tag, latestTag, strings.Join(remoteNames, "\n")})
					}
				}
			}
			return nil
		})
	checkIfError(err)

	if len(rows) == 0 {
		return
	}

	// Print the branch names, tags and remotes in a table
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Repository", "Head", "Branch", "Tag", "Latest Tag", "Remotes"})
	t.AppendRows(rows)
	if color {
		t.SetStyle(table.StyleColoredDark)
	} else {
		t.SetStyle(table.StyleDefault)
	}
	t.Render()
}

func info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func checkIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func isHidden(filename string) bool {

	if runtime.GOOS == "windows" {
		panic(errors.New("Windows is not supported"))
	}
	if filename[0:1] == "." {
		return true
	}
	return false
}

func canRead(info *os.FileInfo) bool {
	m := (*info).Mode()
	if m&(1<<2) != 0 {
		return true
	}
	return false
}

func countLevel(s []byte) int {
	count := int(0)
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			count++
		}
	}
	return count
}

func getChildLevel(basepath, targpath string) (int, error) {
	rel, err := filepath.Rel(basepath, targpath)
	if err != nil {
		return 0, err
	}
	return countLevel([]byte(rel)) + 1, nil
}

func getRemotes(r *git.Repository) ([]string, error) {
	remotes, err := r.Remotes()
	if err != nil {
		return nil, err
	}

	remoteNames := make([]string, len(remotes))
	for i, remote := range remotes {

		remoteNames[i] = getRemoteName(remote)
	}
	return remoteNames, nil
}

func getRemoteName(r *git.Remote) string {
	var url string
	if len(r.Config().URLs) > 0 {
		url = r.Config().URLs[0]
	}
	return fmt.Sprintf("%s\t%s", r.Config().Name, url)
}

func getCurrentBranchAndTag(repository *git.Repository) (string, string, error) {
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

func getCurrentCommit(repository *git.Repository) (string, error) {
	headRef, err := repository.Head()
	if err != nil {
		return "", err
	}
	headSha := headRef.Hash().String()

	return headSha, nil
}

func isClean(repository *git.Repository) (bool, error) {
	w, err := repository.Worktree()
	if err != nil {
		return false, err
	}
	status, err := w.Status()
	if err != nil {
		return false, err
	}
	return status.IsClean(), err
}

func getLatestTag(repository *git.Repository) (string, error) {
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
