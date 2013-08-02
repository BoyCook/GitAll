## About

This is tool for cloning all repositories of a user or organization

## Params

* `{user}` is the account name. This is case sensitive
* `{action}` either clone or update
* `{dir}` this is the target dir, defaults to current dir '.'

## Usage 

	node gitclone.js {user} {action} {dir}

## Example

Clone all the repositories for the user `BoyCook` (https://github.com/BoyCook) into the directory `boycook`:

	node gitclone.js BoyCook clone /Users/boycook/code/boycook

Update all the repositories for the user `BoyCook` (https://github.com/BoyCook) which exist in directory `boycook`:

	node gitclone.js BoyCook update /Users/boycook/code/boycook

# Install from source

Install to `/usr/local/lib/node_modules/gitclone`

	sudo npm install . -g
	