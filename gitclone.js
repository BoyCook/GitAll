var fs = require('fs');
var request = require('request');

function GitClone(user, targetDir) {
	this.user = user;
	this.targetDir = (targetDir ? targetDir : '.');
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

GitClone.prototype.update = function() {
	console.log('Doing update for [%s] on repos in dir [%s]', this.user, this.targetDir);
	this.updateRepos();
};

GitClone.prototype.updateRepos = function() {
	var items = fs.readdirSync(this.targetDir);
	//Read children of targetDir
	for (var i=0,len=items.length; i<len; i++) {
		var item = items[i];
		var repoLocation = this.targetDir + '/' + item;
		if (fs.existsSync(repoLocation) && fs.statSync(repoLocation).isDirectory()) {
			var gitLoc = repoLocation + '/.git';
			if (fs.existsSync(gitLoc) && fs.statSync(gitLoc).isDirectory()) {
				console.log('Found git repos [%s]', item);	
			}
			// console.log('Found dir [%s]', item);
			// var subDirs = fs.readdirSync(this.targetDir + item);
			// for (var x=0,len=subDirs.length; x<len; x++) {
			// 	var subDir = subDirs[x];
			// 	if (fs.statSync(this.targetDir + item + '/' + subDir).isDirectory() && subDir === '.git') {
			// 		console.log('--- Found sub dir [%s]', this.targetDir + item + '/' + subDir);
			// 	}
			// }
		}
	}
};

GitClone.prototype.cloneRepos = function() {
	console.log('Cloning [%s] repositories for user [%s] to [%s]', this.repos.length, this.user, this.targetDir);
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

function exit(msg) {
	console.log(msg);
	process.exit(1);	
}

var args = process.argv.splice(2);

if (args.length < 2) {
	console.log('ERROR - you must pass the correct paramaters, usage:');
	exit('node gitclone.js {username} {action} {dir}')
}

var user = args[0];
var action = args[1];
var dir = args[2];

switch (action) {
	case "clone":
		new GitClone(user, dir).clone();
		break;
	case "update":
		new GitClone(user, dir).update();
		break;	
	default:
		exit('You must use a valid action [clone|update]')
		break;			
}
