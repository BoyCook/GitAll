var fs = require('fs');
var request = require('request');

function GitClone(user) {
	this.user = user;
	this.repos = [];
	this.apiURL = 'https://api.github.com';
}

GitClone.prototype.clone = function() {
	//Doing this nonsense due to dodgy context switch
	var context = this;
	this.getRepos(function() {
		context.cloneRepos();
	});
};

GitClone.prototype.cloneRepos = function() {
	console.log('Cloning [%s] repositories for user [%s]', this.repos.length, this.user);
	for (var i=0,len=this.repos.length; i<len; i++) {
		var repo = this.repos[i];
		console.log('Repos [%s]', repo.name);
	}
};

GitClone.prototype.getRepos = function(success) {
	var context = this;
	var url = this.apiURL + '/users/' + this.user +  '/repos';
	var callBack = function(error, response, body) {
		if (!error && response.statusCode == 200) {
			context.repos = JSON.parse(body);
			if (success) {
				success();
			}
		}
	};
	console.log('Fetching repos from [%s]', url);
	request({url: url, headers: { Accept: 'application/json'}}, callBack);
};

GitClone.prototype.readFile = function(name) {
	this.repos = JSON.parse(fs.readFileSync(name, 'utf8'));
};

var args = process.argv.splice(2);

if (args.length != 1) {
	console.log('ERROR - you must pass the correct paramaters, usage:');
	console.log('node gitclone.js {username} {action}');
	process.exit(1);
}

new GitClone(args[0]).clone();


