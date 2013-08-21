var fs = require('fs');
var util = require('util');
var request = require('request');
var spawn = require('child_process').spawn;
var logLocation = process.env['HOME'] + '/.gitall/gitall.log';

function GitAll(user, targetDir, protocol) {
	this.user = user;
	this.targetDir = targetDir;
	this.protocol = protocol;
	this.repos = [];
	this.baseURLs = {
		ssh: 'git@github.com:' + this.user,
		https: 'https://github.com/' + this.user,
		svn: 'https://github.com/' + this.user
	};
	this.urls = {
		api: 'https://api.github.com',
		ssh: this.baseURLs.ssh + '/%s.git',
		https: this.baseURLs.https + '/%s',
		svn: this.baseURLs.svn + '/%s'
	};	
	this.validate();
	// process.cwd()	
}

GitAll.prototype.validate = function() {
	if (!this.isDir(this.targetDir)) {
		this.exit(util.format('ERROR - Target directory [%s] does not exist for user [%s]', this.targetDir, this.user));
	}
	if (!this.urls.hasOwnProperty(this.protocol)) {
		this.exit(util.format('ERROR - Protocol [%s] is not valid for user [%s]', this.protocol, this.user));
	}
	//TODO: check user account exists
};

GitAll.prototype.clone = function() {
	// Doing this nonsense due to dodgy context switch
	var context = this;
	this.getRepos(function() {
		context.cloneRepos();
	});
};

GitAll.prototype.pull = function() {
	console.log('Doing update for [%s] on repos in dir [%s]', this.user, this.targetDir);
	this.updateRepos();
};

GitAll.prototype.status = function() {
	var repos = this.findRepos();
	for (var i=0,len=repos.length; i<len; i++) {
		this.getRepoStatus(repos[i]);
	}
};

GitAll.prototype.getRepoStatus = function(repo) {
	this.doSpawn('sh', [__dirname + '/status.sh', repo], this.logToConsole);
};

GitAll.prototype.updateRepos = function() {
	this._updateRepos(this.findRepos());
};

GitAll.prototype._updateRepos = function(repos) {
	for (var i=0,len=repos.length; i<len; i++) {
		this.updateRepo(repos[i]);
	}
	console.log('Updated [%s] repos', repos.length);
};

GitAll.prototype.updateRepo = function(repo) {
	this.logToFile(util.format('Updating repo [%s]', repo));
	this.doSpawn('sh', [__dirname + '/pull.sh', repo, this.baseURLs.ssh, this.baseURLs.https], this.logToFile);
};

GitAll.prototype.findRepos = function(callBack) {
	var context = this;
	var items = fs.readdirSync(this.targetDir);
	var repos = [];
	//Read children of targetDir
	for (var i=0,len=items.length; i<len; i++) {
		var item = items[i];
		var repoLocation = this.targetDir + '/' + item;
		var gitLoc = repoLocation + '/.git';
		//Is it a git repo?
		if (this.isDir(gitLoc)) {
			repos.push(repoLocation);
		}
	}
	return repos;
};

GitAll.prototype.isFile = function(name) {
	return (fs.existsSync(name) && fs.statSync(name).isFile());
};

GitAll.prototype.isDir = function(loc) {
	return (fs.existsSync(loc) && fs.statSync(loc).isDirectory());
};

GitAll.prototype.cloneRepos = function() {
	for (var i=0,len=this.repos.length; i<len; i++) {
		this.cloneRepo(this.repos[i]);
	}
	console.log('Cloned [%s] repositories for user [%s] to [%s]', this.repos.length, this.user, this.targetDir);
};

GitAll.prototype.cloneRepo = function(repo) {
	var repoDir = this.targetDir + '/' + repo.name;
	if (this.isDir(repoDir)) {
		this.logToFile(util.format('Not cloning repo [%s] it already exists at [%s]', repo.name, repoDir));
	} else {
		var repoURL = this.generateRepoURL(repo.name);
		this.logToFile(util.format('Cloning repo [%s]', repoURL));
		this.doSpawn('sh', [__dirname + '/clone.sh', this.targetDir, repoURL], this.logToFile);			
	}
};

GitAll.prototype.generateRepoURL = function(repo) {
	return util.format(this.urls[this.protocol], repo);
};

GitAll.prototype.getRepos = function(success) {
	var context = this;
	var url = this.urls.api + '/users/' + this.user +  '/repos?per_page=100';
	var callBack = function(error, response, body) {
		if (!error && response.statusCode == 200) {
			context.repos = JSON.parse(body);
			if (success) {
				success();
			}
		}
		//TODO: on error print message and exit
	};
	console.log('Fetching repos from [%s]', url);
	request({url: url, headers: { Accept: 'application/json'}}, callBack);
};

GitAll.prototype.doSpawn = function(name, args, log) {
    var spawned = spawn(name, args);
    spawned.stdout.on('data', log);
    spawned.stderr.on('data', log);	
};

GitAll.prototype.exit = function(msg) {
	console.log(msg);
	console.log('Usage: gitall {action} {username} {dir}');
	process.exit(1);	
};

GitAll.prototype.logToFile = function(data) {
    fs.appendFileSync(logLocation, data);
};

GitAll.prototype.logToConsole = function (data) {
    console.log(String(data).trim());
};

module.exports = GitAll;
