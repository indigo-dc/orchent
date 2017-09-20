# User Guide
Using orchent is made as easy as possible. In case you are lost orchent provides a lot of
information with its 'help' command, just call `orchent --help`.

## Setting The Access Token
Orchent uses so called access token to authorize itself against the orchestrator.

In The newest release orchent supports the usage of the [oidc-agent](https://github.com/indigo-dc/oidc-agent). By using the oidc-agent the need to copy and paste access tokens is history.
Two things need to be done to use the oidc-agent with orcht. The first thing is to export the
name of the oidc-agent account to use in the environmental variable 'ORCHENT_AGENT_ACCOUNT'.
The account must be loaded into the agent before usage. The second thing is to ensure that
the path to the socket of the oidc-agent is set within the variable 'OIDC_SOCK':

```
export ORCHENT_AGENT_ACCOUNT=<account name>
export OIDC_SOCK=<path to socket of oidc-agent>
```

One can still set the access token directly in the environmental variable 'ORCHENT_TOKEN',
this overrides the previous settings.
`ORCHENT_TOKEN`:
```
export ORCHENT_TOKEN=<your access token here>
```

One can also export the url of the orchestrator via environment variable 'ORCHENT_URL':
```
export ORCHENT_URL=<url to the orchestrator>
```
It is also possible to specify the url at the command line, see below.

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


## Configure Orchent
Orchent supports the ini configuration file format. The configuration file must be located
at `~/.config/orchent/orchent.conf`.

### Alias
Aliases are configured within the 'alias' section. Each line represents an alias and
a uuid to use instead.

An example configuration is:
```
[alias]
one = 78f1a558-fa2d-4415-8bab-058a26d43a79
two = a741b980-3057-48ce-b35a-93af3844953f
```
if such a file exists the alias can be used anywhere instead of the uuid.
The following two commands are performing the same request at the
orchestrator (given the config above):
```
orchent depshow 78f1a558-fa2d-4415-8bab-058a26d43a79
orchent depshow one
```

## Using Orchent
Please make sure you have exported your access token, see above.

### Getting help
orchent provides a lot of help, the main help is shown by running `orchent --help`.
The output is:
```
$ orchent --help
usage: orchent [<flags>] <command> [<args> ...]

The orchestrator client.


Please either store your access token in 'ORCHENT_TOKEN' or set the account to use with oidc-agent
in the 'ORCHENT_AGENT_ACCOUNT' and the socket of the oidc-agent in the 'OIDC_SOCK' environment
variable:

  export ORCHENT_TOKEN=<your access token>
          OR
  export OIDC_SOCK=<path to the oidc-agent socket> (usually this is already exported)
  export ORCHENT_AGENT_ACCOUNT=,account to use>

If you need to specify the file containing the trusted root CAs use the 'ORCHENT_CAFILE' environment
variable:

  export ORCHENT_CAFILE=<path to file containing trusted CAs>


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
    test if the given url is pointing to an orchestrator, please use this to ensure there is no
    typo in the url.

```

the help commang gives even more detailed information on more advanced commands e.g. on 'depcreate':
```
$ orchent help depcreate
usage: orchent depcreate [<flags>] <template> <parameter>

create a new deployment

Flags:
      --help         Show context-sensitive help (also try --help-long and --help-man).
      --version      Show application version.
  -u, --url=URL  the base url of the orchestrator rest interface. Alternative the environment
                 variable 'ORCHENT_URL' can be used: 'export ORCHENT_URL=<the_url>'
      --callback=""  the callback url

Args:
  <template>   the tosca template file
  <parameter>  the parameter to set (json object)
```

### Selecting The Orchestrator
All commands, except with the help command above, need the base url of the orchestrator
to be set. This can be done by setting it in the 'ORCHESTRATOR_URL' environment variable
(see above) or via the url flag.
The flag can be set at any position of the command, yet we recommend setting it as
the first parameter, if not using the environment variable:
```
$ orchent --url=https://my.orchestrator.info/some/path
```

### The Orchent Commands
In this chapter all available commands will be explained.
With two assumptions:
 - The access token is exported
 - The base url is exported

so instead of always adding the url flag like
```
$ orchent --url=https://my.orchestrator.info/some/path <some command>
```
only the command part will be shown like:
```
$ orchent <some command>
```

#### Testing the Orchent URL - test
orchent has a simple way to test if the url points to an orchestrator:
```
$ orchent test
looks like the orchent url is valid
```
The outpt will let you know if the given url looks fine or not. This should be the first
test to perform when having issues, as most of the time a simple typo is the cause of all evil.


#### List Deployments - depls
To list all deployments at an orchestrator the 'depls' command is used.
Invoking is as simple as:
```
$ orchent depls
```
The output is a long list of all pages of deployments.

The output can be filtered to a specific user by adding the subject@issuer:
```
$ orchent depls --created_by=som-uuid-at-iam@https://iam-test.indigo-datacloud.eu/
```
There is also a shortcut for the current user - 'me':
```
$ orchent depls --created_by=me
```

It is also possible to filter the deployments by data/time. The flags used are `--before` and
`--after`, they can be used either alone or together as well as in combination with `--created_by`.
A date/time is specified as 'YYYYMMDDHHMM'. The correctness of the date or time is not checked.
```
$ orchent depls --after=201707090000
```


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
$ orchent depcreate ambertools.yaml '{ "number_cpus": 1, ​"memory_size": "1 GB" }'
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
an example template would be the [ambertools template](https://raw.githubusercontent.com/indigo-dc/tosca-templates/master/amber/ambertools.yaml)

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
