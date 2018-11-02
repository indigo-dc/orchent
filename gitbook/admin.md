# Deployment And Administration Guide
## Installation
### From Package
For Ubuntu 14.04 and CentOS 7 a package is provided by the INDIGO DataCloud
team.
To be able to install the packages using the package manager of your system, the
repository needs to be added. This is done by adding the INDIGO DataCloud
package repository to your system.

The INDIGO DataCloud repository can be found at http://repo.indigo-datacloud.eu .


For informations on how to add the repository to your system refer to the
documentation of your operating system.

#### Ubuntu 14.04
After adding the repository one needs to update the package list and then install
the Token Translation Service.
```
apt update
apt install orchent
```

#### CentOS 7
After adding the repository one needs to update the package list and then install
the Token Translation Service.
```
yum update
yum install orchent
```

### From Source
To be able to install orchent from source, you need the go programming language
version 1.5 or newer installed on the system.

The go programming language should be in the repository of your Linux distribution.

On Debian/Ubuntu as well as on CentOs it is called 'golang'.
So installing go on Debian based distributions is just a
```
apt install golang
```

Installing on CentOs is also just a
```
yum install golang
```

#### Importing the Docker Container
If your system is not supported you can still use orchent through a lightweight Docker container.
Download the container in the [release section](https://github.com/indigo-dc/orchent/releases) (choose the latest stable version) and import it using the `docker load` command, e.g.:
```
docker load --input orchent_container_1.0.4.tar
```

After loading the container you can use it to run orchent:
```
docker run orchent:1.0.4 --version
docker run orchent:1.0.4 --help
```

For information on how to pass environment settings to the docker see
```
docker run --help
```

#### The official go way
Orchent is following the
[go code organization rules](https://golang.org/doc/code.html#Organization)
and should be placed in the `src/github.com/indigo-dc/orchent` sub-directory
of the `$GOPATH`. So if you have a working go-lang setup, orchent can be easily
build using `go build`.

#### The Fast And Easy Way
For anyone just wanting to compile orchent without setting up a complete go-lang development
environment; The next steps build download and build orchent:
- clone the git repository
- build the package yourself
```
git clone https://github.com/indigo-dc/orchent
cd orchent
./utils/compile.sh
```
practically it creates the go development environment for you.
The binary executable is in the current direcotry called 'orchent'.

One could now copy the binary to e.g. `/usr/bin`.
