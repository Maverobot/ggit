[![Build Status][github-actions-badge]][github-actions-link]
[![MIT License][license-badge]](LICENSE)
[![Codacy Badge][codacy-badge]][codacy-link]
# ggit
A simple program to list status of all repositories under a given directory

## Getting Started

Download the binary and use the program right away:

As example for amd64 in Linux:

```sh
wget -qc https://github.com/Maverobot/ggit/releases/download/v0.2.3/ggit_0.2.3_linux_amd64.tar.gz -O - | tar -C ~/.local/bin/ -xz ggit
chmod +x ~/.local/bin/ggit
```

## Usage
```sh
Usage: ggit [flags]
  -color
    	Whether the table should be rendered with color. (default true)
  -depth int
    	The depth ggit should go searching. (default 2)
  -path string
    	The path to the parent directory of git repos. (default "./")
  -update
    	Try go-github-selfupdate via GitHub
  -version
    	Show version
```

Example:
```sh
# By default, it takes current directory path as input
cd a_folder_with_many_repos
ggit

# Or,
ggit -path path_to_folder_with_many_repos -depth 1
```

Simple showcase:

![](demo.gif)

[github-actions-badge]: https://github.com/maverobot/ggit/actions/workflows/build.yaml/badge.svg?branch=master
[github-actions-link]: https://github.com/Maverobot/ggit/actions
[codacy-badge]:     https://api.codacy.com/project/badge/Grade/840d280344b245a38ed80cecf38cf96b
[codacy-link]:      https://www.codacy.com/manual/quzhengrobot/ggit?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=Maverobot/ggit&amp;utm_campaign=Badge_Grade
[license-badge]:    https://img.shields.io/badge/License-MIT-blue.svg
[release-download]: https://github.com/Maverobot/ggit/releases/download/continuous/ggit-linux-amd64
