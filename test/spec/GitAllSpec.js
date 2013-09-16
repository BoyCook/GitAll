var should = require('should');
var GitAll = require('../../lib/gitall');

describe('GitAll', function () {
    it('should validate ok', function () {
        new GitAll('user', './', 'ssh');
    });
});
