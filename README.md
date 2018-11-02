# orchent
The Orchestrator Command Line Client

## Building orchent

Before compiling orchent, make sure to install [The Go Programming Language](https://golang.org)

```
# Building orchent
git clone https://github.com/indigo-dc/orchent.git
cd orchent/utils

# Linux
./compile.sh

# macOS
./compile_macos.sh

# Test the installation
./orchent --help
```

## Usage
orchent helps you as much as possible:
```
usage: orchent [<flags>] <command> [<args> ...]

The orchestrator client. Please store your access token in the 'ORCHENT_TOKEN' environment
variable: 'export ORCHENT_TOKEN=<your access token>'. If you need to specify the file
containing the trusted root CAs use the 'ORCHENT_CAFILE' environment variable:
'export ORCHENT_CAFILE=<path to file containing trusted CAs>'.

Flags:
      --help     Show context-sensitive help (also try --help-long and --help-man).
      --version  Show application version.
  -u, --url=URL  the base url of the orchestrator rest interface. Alternative the environment
                 variable 'ORCHENT_URL' can be used: 'export ORCHENT_URL=<the_url>'

Commands:
  help [<command>...]
    Show help.

  depls [<flags>]
    list deployments

  depshow <uuid>
    show a specific deployment

  depcreate [<flags>] <template> <parameter>
    create a new deployment

  depupdate [<flags>] <uuid> <template> <parameter>
    update the given deployment

  deptemplate <uuid>
    show the template of the given deployment

  depdel <uuid>
    delete a given deployment

  resls <depployment uuid>
    list the resources of a given deployment

  resshow <deployment uuid> <resource uuid>
    show a specific resource of a given deployment

  test
    test if the given url is pointing to an orchestrator, please use this to ensure
    there is no typo in the url.



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


## using Docker
If your system is not supported you can still use orchent through a lightweight Docker container.
Download the container in the [release section](https://github.com/indigo-dc/orchent/releases)(choose the latest stable version) and import it using the `docker load` command, e.g.:
```
docker load --input orchent_container_1.1.0.tar
```

After loading the container you can use it to run orchent:
```
docker run orchent:1.1.0 --version
docker run orchent:1.1.0 --help
```

For information on how to pass environment settings to the docker see
```
docker run --help
```
