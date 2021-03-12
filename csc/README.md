# Container Storage Client
The Container Storage Client (`csc`) is a command line interface (CLI) tool
that provides analogues for all of the CSI RPCs.

```bash
$ csc
NAME
    csc -- a command line container storage interface (CSI) client

SYNOPSIS
    csc [flags] CMD

AVAILABLE COMMANDS
    controller
    identity
    node

Use "csc -h,--help" for more information
```

## Installation

```bash
$ GO111MODULE=off go get -u github.com/dell/gocsi/csc
```
