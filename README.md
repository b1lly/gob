gob
===

Go Language build tool that automatically rebuilds and runs 
the program when any files/packages are added/modified/deleted.

### Go Language Build Utility

This is a very basic utility that wraps around the default 
go build tools to build and run Go programs. 

It basically watches the file system (recursively) based on the 
programs root path and monitors it for any additions, modifications or 
deletions of files. If it detects anything, it simply kills
the existing process and starts up a new one.

### Installation

    go get github.com/b1lly/gob
    go install $GOPATH/src/github.com/b1lly/gob

### Usage

    gob $GOPATH/src/to/myApp.go
    (e.g. gob $GOPATH/src/github.com/b1lly/gob/test/test.go)

This would build and run the `$GOPATH/src/*` for any change and rebuild/run
when a modification event is received.
