## v1.2.7

This version allows to interact with the latest version of the PaaS Orchestrator (v2.6.0), introducing the new command `deplog` and the new option `usergroup` for deployment operations.  
 
### Added

- `deplog` command: this new command allows to retrieve the deployment log if available
 
### Changed

- commands `depcreate` and `depls` provide a new option, `--user_group`, that allows to specify the user group to filter the active deployments or to create a new deployment, respectively.  
- changed build and packaging strategy - github actions setup 

