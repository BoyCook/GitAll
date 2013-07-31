var fs = require('fs');

function GitClone() {
	this.repos = JSON.parse(fs.readFileSync('./data/repos.json', 'utf8'));
	this.apiURL = 'https://api.github.com/users/boycook/gists';
}

GitClone.prototype.clone = function() {
	for (var i=0,len=this.repos.length; i<len; i++) {
		var repo = this.repos[i];
		console.log('Repos [%s]', repo["name"]);
	}
};




new GitClone().clone();
