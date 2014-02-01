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
    
### Basic Usage

    gob $GOPATH/src/to/myApp.go
    (e.g. gob $GOPATH/src/github.com/b1lly/gob/test/test.go)

This would build and run the program and listen to `$GOPATH/src/github.com/b1lly/gob/test*` 
for any change. It then rebuilds and runs the program if a modification 
event is received.

### Advanced CLI Flags

    --agent=[false,true]
    --port=[9034]

### Gob Agent Overview

    The gob/agent package provides a way for your application to talk to
    the gob builder. You simply register a handler with Gob Agent and
    start it up. That will automatically setup a communication channel with the GobServer
    which is built into Gob. GobServer will notifiy GobAgent of template files
    that were changed and will call the handlers that you have registered on the Agent.
    This is very useful for applications that have their own templating engines.
    It provides a way to re-render templates without having to rebuild the entire application.
    
    NOTE: We currently only recognize *.soy templates as template files, but will provide
    a way in the future to customize this.
