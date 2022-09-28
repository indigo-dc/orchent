# orchent

![Build workflow](https://github.com/maricaantonacci/orchent/actions/workflows/main.yaml/badge.svg)

The Orchestrator Command Line Client

## Build from source

Before compiling orchent, make sure to install [The Go Programming Language](https://golang.org)

```
# Building orchent
git clone https://github.com/indigo-dc/orchent.git
cd orchent
go build -o orchent orchent.go

# Test it
./orchent --version
```

## Install package on Linux

Download the rpm/deb package from [here](releases) and install it:

````
# Debian/Ubuntu:
dpkg -i orchent_1.2.7-rc1_amd64.deb 

# CentOS/Fedora
rpm -i orchent-1.2.7-1.el7.x86_64.rpm
````
Test the installation:

````
orchent --version
````

## Install orchent binary on MacOS

Download the darwin package from [here](releases), rename it (optional) and add executable permissions:

````
mv orchent-amd64-darwin orchent
chmod +x orchent
````
Test it:

````
./orchent --version
````

## Usage
orchent helps you as much as possible:
```
usage: orchent [<flags>] <command> [<args> ...]

The orchestrator client.



Please either store your access token in 'ORCHENT_TOKEN' or set the account to use with oidc-agent in the 'ORCHENT_AGENT_ACCOUNT' and the socket of the oidc-agent in the 'OIDC_SOCK' environment variable:

  export ORCHENT_TOKEN=<your access token>
          OR
  export OIDC_SOCK=<path to the oidc-agent socket> (usually this is already exported)
  export ORCHENT_AGENT_ACCOUNT=<account to use>

If you need to specify the file containing the trusted root CAs use the 'ORCHENT_CAFILE' environment variable:

  export ORCHENT_CAFILE=<path to file containing trusted CAs>


Flags:
      --help     Show context-sensitive help (also try --help-long and --help-man).
      --version  Show application version.
  -u, --url=URL  the base url of the orchestrator rest interface. Alternative the environment variable 'ORCHENT_URL' can be used: 'export ORCHENT_URL=<the_url>'

Commands:
  help [<command>...]
    Show help.

  depls [<flags>]
    list deployments

  depshow [<flags>] <uuid>
    show a specific deployment

  depcreate [<flags>] <template> <parameter>
    create a new deployment

  depupdate [<flags>] <uuid> <template> <parameter>
    update the given deployment

  deptemplate <uuid>
    show the template of the given deployment

  depdel <uuid>
    delete a given deployment

  depreset [<flags>] <uuid>
    reset the state of a given deployment

  resls <deployment uuid>
    list the resources of a given deployment

  resshow <deployment uuid> <resource uuid>
    show a specific resource of a given deployment

  test
    test if the given url is pointing to an orchestrator, please use this to ensure there is no typo in the url.

  showconf
    list the endpoints used by the current orchestrator.
```

Before using the orchestrator with orchent you need to export your IAM access token:
```
export ORCHENT_TOKEN=<your access token here>
```

As long as the access token is valid orchent can tell the orchestrator what to do.
e.g. update a deployment:
```
export ORCHENT_URL=https://orchestrator01-indigo.cloud.ba.infn.it/orchestrator/
./orchent depupdate eac4dabb-9613-4026-bac7-6075050308e3 template.txt '{"number_cpus": +1, "memory_size": "2 GB"}'
update of deployment eac4dabb-9613-4026-bac7-6075050308e3 successfully triggered
```
And after that one could e.g. have a look at the deployment:
```
./orchent depshow eac4dabb-9613-4026-bac7-6075050308e3
Deployment [eac4dabb-9613-4026-bac7-6075050308e3]:
  status: UPDATE_IN_PROGRESS
  creation time: 2016-10-12T07:02+0000
  update time: 2016-10-12T07:13+0000
  callback:
  output: map[]
  links:
    self [http://orchestrator01-indigo.cloud.ba.infn.it:8080/orchestrator/deployments/eac4dabb-9613-4026-bac7-6075050308e3]
    resources [http://orchestrator01-indigo.cloud.ba.infn.it:8080/orchestrator/deployments/eac4dabb-9613-4026-bac7-6075050308e3/resources]
    template [http://orchestrator01-indigo.cloud.ba.infn.it:8080/orchestrator/deployments/eac4dabb-9613-4026-bac7-6075050308e3/template]
```
For more information and more examples please see the [documentation](https://indigo-dc.gitbooks.io/orchent/)


