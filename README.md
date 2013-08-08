## About

This is tool for cloning and updating all the GitHub repositories of a user or organization in one hit.

## Params

* `{user}` is the account name. This is case sensitive
* `{action}` either `clone`, `update` or `status`
* `{dir}` this is the target dir, defaults to current dir '.'

## Actions

* `clone` - clones all repositories
* `update` - updates all repositories
* `status` - gives status for all repositories
* `config` - gives the config

## config.json

You can set some paramaters in the file `$HOME/.gitall/config.json` to save using them on the command line. The tool will try and read paramters from this file (if one exists), and will use those if none are passed in. Parameters that can be set are:

	{
	   "username": "{user}",
	   "dir": "{dir}"
	}

## Usage 

	gitall {action} {user} {dir}

The final two are optional

## Example

Clone all the repositories for the user `BoyCook` (https://github.com/BoyCook) into the directory `boycook`:

	gitall clone BoyCook /Users/boycook/code/boycook

Update all the repositories for the user `BoyCook` (https://github.com/BoyCook) which exist in directory `boycook`:

	gitall update BoyCook /Users/boycook/code/boycook

If the `username` and `dir` are set in `$HOME/.gitall/config.json`, these commands become:

	gitall clone
	gitall update
	gitall status

# Install from source

Install to `/usr/local/lib/node_modules/gitall`

	sudo npm install . -g

Or use script

	sudo ./install.sh
	