# rmdupes

rmdupes is a very simple command-line utility to remove duplicate files.

## Usage

Compile or download an executable and put it on your PATH.
Then open up a terminal and enter: `rmdupes <directory>`.
This will remove all duplicate files in that directory, based on their CRC32 checksum.
As this tool can potentially remove very many files, running it in the current working directory requires you to name the directory explicitely:
`rmdupes .`

## Compile

Install the go tool-chain on your system, clone this repo, navigate to it and enter `go build` in your terminal.

## Motivation

This is my first publish-worthy tool I have written in GO.
I wrote it, as I was looking for a tool doing exactly this.
Furthermore, I wanted to get some hands-on programming experience with this language and especially with it's concurrency features.

