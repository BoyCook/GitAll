## About

This is a tool to mange (clone, pull etc) all the GitHub repositories for multiple user (or organisation) accounts in one command. 

Do you work with multiple GitHub repositories over multiple user or organisation accounts? Ever wanted to clone or update all your GitHub repositories with one command? This is the tool for you.

## Params

* `{action}` either [clone|pull|status|config]
* `{user}` is the account name. This is case sensitive
* `{dir}` this is the target dir, defaults to current dir '.'
* `{protocol}` [ssh|https|svn] this is the protocol to be used to fetch the repo, defaults to 'ssh' 

## Actions

* `clone` - clones all repositories
* `pull` - updates all repositories
* `status` - gives status for all repositories
* `config` - gives the config in `$HOME/.gitall/config.json`

## config.json

You setup the config for the accounts that you want to manage in a config file `$HOME/.gitall/config.json`.
Example config is:

	[{
	   "username": "BoyCook",
	   "dir": "/Users/boycook/code/boycook",
	   "protocol": "ssh"
	},{
	   "username": "TiddlySpace",
	   "dir": "/Users/boycook/code/osmosoft/tiddlyspace",
	   "protocol": "ssh"
	}]

## How it works

GitAll works by either setting up config in the config file, or passing it parameters on the command line. 
Parameters passed in will take precidence over parameters found in the config file. 
It's much better to setup the config in advance and let the GitAll do all the hard work.

## Usage 

	gitall {action} {user} {dir} {protocol}

The final three are optional

## Example usage with config file

	gitall clone
	gitall pull
	gitall status

These will perform the action specified on each account setup in the config file.

## Example usage passing in parameters

	gitall clone BoyCook /Users/boycook/code/boycook ssh

This will clone all the repositories for the user `BoyCook` (https://github.com/BoyCook) into the directory `boycook` using 
the `ssh` protocol.

# Install from source

Install to `/usr/local/lib/node_modules/gitall`

	sudo npm install . -g

Or use script

	sudo ./install.sh
	
## Prerequisites

* GitAll is a node.js app so http://nodejs.org will be required.
* You may want to increase the number of file descriptors allowed `ulimit -n 10000`
	