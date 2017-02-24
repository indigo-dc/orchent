# User Guide
Using orchent is made as easy as possible. In case you are lost orchent provides a lot of
information with its 'help' command, just call `orchent --help`.

## Setting The Access Token
The orchestrator needs a way to authorize orchent, this is done by a so called access token.
The access token is retrieved beforhand at either [IAM](https://github.com/indigo-iam/iam) or
[TTS](https://github.com/indigo-dc/tts).

Once an access token is known, it needs to be exportet in the environment variable
`ORCHENT_TOKEN`:
```
export ORCHENT_TOKEN=<your access token here>
```

Now the orchent can perform any operation the access token grants, as long as the access token
is valid.

## Setting the trusted Certificate Authorities (CAs)
Usually this part is not needed as most systems come with a sane default setup.

Sometimes you either need or want to specify which CAs you trust anyway.
You can explicitly tell orchent which file contains all the root CAs that can be
trusted by using the `ORCHENT_CAFILE` environment variable. The file must contain
the certficates in the PEM format.
```
export ORCHENT_CAFILE=<path to the file containing trusted CAs>
```


## Using Orchent
Please make sure you have exported your access token, see above.

### Getting help
orchent provides a lot of help, the main help is shown by running `orchent help`.
The output is:
```
$ orchent help
usage: orchent --url=URL [<flags>] <command> [<args> ...]

The orchestrator client. Please store your access token in the 'ORCHENT_TOKEN' environment
variable: 'export ORCHENT_TOKEN=<your access token>'. If you need to specify the file
containing the trusted root CAs use the 'ORCHENT_CAFILE' environment variable:
'export ORCHENT_CAFILE=<path to file containing trusted CAs>'.

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

the help commang gives even more detailed information on more advanced commands e.g. on 'depcreate':
```
$ orchent help depcreate
usage: orchent depcreate [<flags>] <template> <parameter>

create a new deployment

Flags:
      --help         Show context-sensitive help (also try --help-long and --help-man).
      --version      Show application version.
  -u, --url=URL      the base url of the orchestrator rest interface
      --callback=""  the callback url

Args:
  <template>   the tosca template file
  <parameter>  the parameter to set (json object)
```

### Selecting The Orchestrator
All commands, except with the help command above, need the 'url' flag to be set.
The url flag defines the base url of the orchestrator to connect to.
The flag can be set at any position of the command, yet we recommend settint it as
the first parameter:
```
$ orchent --url=https://my.orchestrator.info/some/path
```

### The Orchent Commands
In this chapter all available commands will be explained.
With two assumptions:
 - The access token is exported
 - The base url is known and will be added to the command

so instead of always adding the url flag like
```
$ orchent --url=https://my.orchestrator.info/some/path <some command>
```
only the command part will be shown like:
```
$ orchent <some command>
```

#### List All Deployments - depls
To list all deployments at an orchestrator the 'depls' command is used.
Invoking is as simple as:
```
$ orchent depls
```
The output is a long list of all pages of deployments.


#### Show A Specific Deployment - depshow
To show only a specific deployment orchent needs the uuid of the deployment.
```
$ orchent depshow <uuid>
```
Example:
```
$ orchent depshow 12345678-1234-1234-1234-123456789abc
Deployment [12345678-1234-1234-1234-123456789abc]:
  status: CREATE_COMPLETE
  status reason:
  creation time: 2016-07-22T11:08+0000
  update time: 2016-07-22T11:09+0000
  callback:
  output: map[]
  links:
    self [./deployments/12345678-1234-1234-1234-123456789abc]
    resources [./deployments/12345678-1234-1234-1234-123456789abc/resources]
    template [./deployments/12345678-1234-1234-1234-123456789abc/template]

```
#### Get The Template Of A Deployment - deptemplate
To get the template of a deployment the 'deptemplate' command is used:
```
orchent deptemplate <uuid>
```
Example:
```
$ orchent deptemplate 12345678-1234-1234-1234-123456789abc

tosca_definitions_version: tosca_simple_yaml_1_0

imports:
  - indigo_custom_types: https://raw.githubusercontent.com/indigo-dc/tosca-types/master/custom_types.yaml

description: >
  TOSCA examples for specifying a Chronos Job that runs an application using Onedata storage.

topology_template:
  inputs:
    input_onedata_token:
[ .... ]
```

#### Create A New Deployment - depcreate
The creation of a new deployment is done using the 'depcreate' command, it needs two parameter:
 - the name of the template file
 - a string containing the JSON object with parameter, best is to use

```
$ orchent depcreate <template file name> <json object with parameter>
```

Example:
```
$ orchent depcreate ../myTemplate.yml '{ "number_cpus": 1, ​"memory_size": "1 GB" }'
Deployment [12345678-1234-1234-1234-123456789abc]:
  status: CREATE_IN_PROGRESS
  status reason:
  creation time: 2016-09-09T:07+0000
  update time:
  callback:
  output: map[]
  links:
    self [./deployments/12345678-1234-1234-1234-123456789abc]
    resources [./deployments/12345678-1234-1234-1234-123456789abc/resources]
    template [./deployments/12345678-1234-1234-1234-123456789abc/template]
```
#### Update A Given Deployment - depupdate
Updating a deployment is like creating one with the difference that a uuid of an existing
deployment needs to be passed:
```
$ orchent depupdate <uuid of the deployment to update> <template file name> <json parameter>
```
Example:
```
$ orchent depupdate 12345678-1234-1234-1234-123456789abc ../myTemplate.yml '{ "number_cpus": 2, ​"memory_size": "2 GB" }'
update of deployment 12345678-1234-1234-1234-123456789abc successfully triggered
```
#### Delete A Given Deployment - depdel
To delete a deployment only its uuid is needed:
```
$ orchent depdel <uuid>
```
Example:
```
$ orchent depdel 12345678-1234-1234-1234-123456789abc
deletion of deployment 12345678-1234-1234-1234-123456789abc successfully triggered
```
#### Get The Resources Of A Deployment - resls
Listing all the resources of a deployment is very similar to listing all deployments,
only that a deployment uuid must be passed:
```
$ orchent resls <deployment uuid>
```
Example:
```
$ orchent resls 12345678-1234-1234-1234-123456789abc
```
The result is a long list of resources for the given deployment

#### Show A Specific Resource Of A Deployment - resshow
To only display one specific resource of a given deployment 'resshow' needs the uuid of both of them,
the deployment and the resource:
```
$ orchent resshow <deployment uuid> <resource uuid>
```
Example:
```
$ orchent resshow 12345678-1234-1234-1234-123456789abc 99999999-9999-9999-9999-999999999999
Resource [99999999-9999-9999-9999-999999999999]:
  creation time: 2016-07-22T11:08+0000
  state: STARTED
  toscaNodeType: tosca.nodes.indigo.Container.Application.Docker.Chronos
  toscaNodeName: chronos_job
  requiredBy:
  links:
    deployment [./deployments/12345678-1234-1234-1234-123456789abc]
    self [./deployments/12345678-1234-1234-1234-123456789abc/resources/99999999-9999-9999-9999-999999999999]
```
