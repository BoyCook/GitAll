
TESTS = test/spec
REPORTER = spec
XML_FILE = reports/TEST-all.xml
HTML_FILE = reports/coverage.html

test: test-mocha

test-ci:
	$(MAKE) test-mocha REPORTER=xUnit > $(XML_FILE)

test-all: clean test-ci test-cov

test-ui: start
	casperjs test test/ui

test-mocha:
	@NODE_ENV=test mocha \
	    --timeout 200 \
		--reporter $(REPORTER) \
		$(TESTS)

test-cov: lib-cov
	@HFS_COV=1 $(MAKE) test-mocha REPORTER=html-cov > $(HTML_FILE)

lib-cov: setup
	jscoverage lib lib-cov

setup:
	mkdir -p reports

clean:
	rm -f reports/*
	rm -fr lib-cov
