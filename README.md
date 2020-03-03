[![Build Status][travis-badge]][travis-link]
[![MIT License][license-badge]](LICENSE)
# ggit
A simple program to list status of all repositories under a given directory

## Getting Started

Download the [AppImage][release-download] and use the program right away:

```sh
wget https://github.com/Maverobot/ggit/releases/download/continuous/ggit-continuous.glibc2.4-x86_64.AppImage -O ~/.local/bin/ggit
chmod +x ~/.local/bin/ggit

# By default, it takes current directory path as input
cd a_folder_with_many_repos
ggit

# Or,
ggit path_to_folder_with_many_repos
```


[travis-badge]:     https://travis-ci.com/Maverobot/ggit.svg?branch=master
[travis-link]:      https://travis-ci.com/Maverobot/ggit
[license-badge]:    https://img.shields.io/badge/License-MIT-blue.svg
[release-download]: https://github.com/Maverobot/ggit/releases/download/continuous/ggit-continuous.glibc2.4-x86_64.AppImag
