# orchent
The Orchestrator Command Line Client

## Usage
orchent helps you as much as possible:
```
$ orchent --help
usage: orchent [<flags>] <command> [<args> ...]

The orchestrator client. Please store your access token in the 'ORCHENT_TOKEN'
environment variable: 'export ORCHENT_TOKEN=<your access token>'. If you need to
specify the file containing the trusted root CAs use the 'ORCHENT_CAFILE'
environment variable: 'export ORCHENT_CAFILE=<path to file containing trusted
CAs>'.

Flags:
      --help     Show context-sensitive help (also try --help-long and
                 --help-man).
      --version  Show application version.
  -u, --url=URL  the base url of the orchestrator rest interface. Alternative
                 the environment variable 'ORCHENT_URL' can be used: 'export
                 ORCHENT_URL=<the_url>'

Commands:
  help [<command>...]
    Show help.

  depls
    list all deployments

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


```

Before using the orchestrator with orchent you need to export your IAM access token:
```
export ORCHENT_TOKEN=<your access token here>
```

As long as the access token is valid orchent can tell the orchestrator what to do.
e.g. update a deployment:
```
./orchent --url=http://orchestrator01-indigo.cloud.ba.infn.it:8080/orchestrator depupdate eac4dabb-9613-4026-bac7-6075050308e3 template.txt '{"number_cpus":
+1, "memory_size": "2 GB"}'
update of deployment eac4dabb-9613-4026-bac7-6075050308e3 successfully triggered
```
And after that one could e.g. have a look at the deployment:
```
./orchent --url=http://orchestrator01-indigo.cloud.ba.infn.it:8080/orchestrator depshow eac4dabb-9613-4026-bac7-6075050308e3
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
Download the container in the release section and import it using the `docker load` command:
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
