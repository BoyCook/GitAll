var fs = require('fs');
var request = require('request');
var spawn = require('child_process').spawn;

function GitClone(user, targetDir) {
	this.user = user;
	this.targetDir = (targetDir ? targetDir : '.');
	this.repos = [];
	this.apiURL = 'https://api.github.com';
	this.repoURLBase = 'git@github.com:' + this.user + '/';
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
	this._updateRepos(this.findRepos());
	// process.cwd()
};

GitClone.prototype._updateRepos = function(repos) {
	for (var i=0,len=repos.length; i<len; i++) {
		this.updateRepo(repos[i]);
	}
	console.log('Updated [%s] repos', repos.length);
};

GitClone.prototype.updateRepo = function(repo) {
	console.log('Updating repo [%s]', repo);			
	this.doSpawn('sh', ['pull.sh', repo]);
};

GitClone.prototype.findRepos = function() {
	var items = fs.readdirSync(this.targetDir);
	var repos = [];
	//Read children of targetDir
	for (var i=0,len=items.length; i<len; i++) {
		var item = items[i];
		var repoLocation = this.targetDir + '/' + item;
		if (fs.existsSync(repoLocation) && fs.statSync(repoLocation).isDirectory()) {
			var gitLoc = repoLocation + '/.git';
			if (fs.existsSync(gitLoc) && fs.statSync(gitLoc).isDirectory()) {
				repos.push(repoLocation);
			}
		}
	}
	return repos;
};

GitClone.prototype.cloneRepos = function() {
	for (var i=0,len=this.repos.length; i<len; i++) {
		this.cloneRepo(this.repos[i]);
	}
	console.log('Cloned [%s] repositories for user [%s] to [%s]', this.repos.length, this.user, this.targetDir);
};

GitClone.prototype.cloneRepo = function(repo) {
	var repoURL = this.repoURLBase + repo.name + '.git';
	console.log('Cloning repo [%s]', repoURL);			
	this.doSpawn('sh', ['clone.sh', this.targetDir, repoURL]);	
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

GitClone.prototype.doSpawn = function(name, args) {
	var log = function(data) {
		console.log(String(data));
	};
    var spawned = spawn(name, args);
    spawned.stdout.on('data', this.logToFile);
    spawned.stderr.on('data', this.logToFile);	
};

GitClone.prototype.logToFile = function(data) {
    fs.appendFileSync('/Users/boycook/.gitclone/gitclone.log', data, function (err) {
        console.log('Error writing to file');
        if (err) throw err;
    });
}

GitClone.prototype.logToConsole = function (data) {
    console.log(String(data));
}

module.exports = GitClone;
