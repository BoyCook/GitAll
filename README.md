## About

This is tool for cloning and updating all the GitHub repositories of a user or organization in one hit.

## Params

* `{user}` is the account name. This is case sensitive
* `{action}` either `clone`, `update` or `status`
* `{dir}` this is the target dir, defaults to current dir '.'

## config.json

You can set some paramaters in the file `$HOME/.gitclone/config.json` to save using them on the command line. The tool will try and read paramters from this file (if one exists), and will use those if none are passed in. Parameters that can be set are:

	{
	   "username": "{user}",
	   "dir": "{dir}"
	}

## Usage 

	gitclone {action} {user} {dir}

The final two are optional

## Example

Clone all the repositories for the user `BoyCook` (https://github.com/BoyCook) into the directory `boycook`:

	gitclone clone BoyCook /Users/boycook/code/boycook

Update all the repositories for the user `BoyCook` (https://github.com/BoyCook) which exist in directory `boycook`:

	gitclone update BoyCook /Users/boycook/code/boycook

If the `username` and `dir` are set in `$HOME/.gitclone/config.json`, these commands become:

	gitclone clone
	gitclone update
	gitclone status

# Install from source

Install to `/usr/local/lib/node_modules/gitclone`

	sudo npm install . -g
	