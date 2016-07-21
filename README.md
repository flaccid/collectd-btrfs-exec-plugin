# collectd-btrfs-exec-plugin

Inspired from https://github.com/soellman/docker-collectd/blob/master/btrfs-data.py.

## Usage
TODO

## Build

    $ go install github.com/flaccid/collectd-btrfs-exec-plugin

Or, build the binary locally within the repository folder:

    $ go build exec-btrfs.go

## Cross-compiling

Of course, BTRFS is mostly native to Linux, but perhaps you can FUSE or similar.
See https://golang.org/doc/install/source#environment for possibilities.

Mac OS X:

    $ GOOS=darwin GOARCH=amd64 go build exec-btrfs.go

Linux:

    $ GOOS=linux GOARCH=amd64 go build exec-btrfs.go

Microsoft Windows:

    $ GOOS=windows GOARCH=amd64 go build exec-btrfs.go

## Upstream Resources
https://collectd.org/wiki/index.php/Plugin:Exec
https://collectd.org/wiki/index.php/Category:Exec_Plugins
