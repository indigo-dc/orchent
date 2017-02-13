# orchent
The Orchestrator Command Line Client

## Usage
orchent helps you as much as possible:
```
./orchent --help
usage: orchent --url=URL [<flags>] <command> [<args> ...]

The orchestrator client. Please store your access token in the 'ORCHENT_TOKEN' environment variable: 'export ORCHENT_TOKEN=<your access token>'

Flags:
      --help     Show context-sensitive help (also try --help-long and --help-man).
      --version  Show application version.
  -u, --url=URL  the base url of the orchestrator rest interface

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
