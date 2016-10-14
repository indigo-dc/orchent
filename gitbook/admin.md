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

The next steps build download and build orchent:
- clone the git repository
- build the package yourself
- install the package
```
git clone https://github.com/indigo-dc/orchent
cd orchent
make
```
the binary executable is in the current direcotry called 'orchent'.

One could now copy the binary to e.g. `/usr/bin`.
