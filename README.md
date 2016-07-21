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

- https://collectd.org/wiki/index.php/Plugin:Exec
- https://collectd.org/wiki/index.php/Category:Exec_Plugins

License and Authors
-------------------
- Author: Chris Fordham (<chris@fordham-nagy.id.au>)

```text
Copyright 2016, Chris Fordham

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
