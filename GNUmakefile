ARCH := amd64
VERSION := 1.0.2

TRAVIS_BUILD_NUMBER ?= nontravis
PKGNAME := growse-com-locator-and-search
TEST_REPORT := test-reports/report.xml
TEST_COVERAGE := test-reports/coverage.html

GO := /usr/lib/go-$(GOVERSION)/bin/go

.PHONY: package

package: $(addsuffix .deb, $(addprefix $(PKGNAME)_$(VERSION)-$(TRAVIS_BUILD_NUMBER)_, $(foreach a, $(ARCH), $(a))))

$(PKGNAME)_$(VERSION)-$(TRAVIS_BUILD_NUMBER)_%.deb: dist/www-growse-com_linux_%
	bundle exec fpm -f -s dir -t deb --url https://www.growse.com/ --description "growse.com dynamic content (locator, search)" --deb-systemd www-growse-com.service -n $(PKGNAME) --config-files /etc/www-growse-com.conf -p . -a $* -v $(VERSION)-$(TRAVIS_BUILD_NUMBER) $<=/usr/bin/www.growse.com config.json=/etc/www-growse-com.conf databasemigrations/=/var/www/growse-web/databasemigrations templates/=/var/www/growse-web/templates

$(GOPATH)/bin/go-junit-report:
	$(GO) get -u github.com/jstemmer/go-junit-report

.PHONY: test
test: $(TEST_COVERAGE)

$(TEST_REPORT): $(GOPATH)/bin/go-junit-report
	mkdir -p test-reports
	$(GO) test -cover -covermode=count -coverprofile=test-reports/coverprofile -v | tee /dev/tty | $(GOPATH)/bin/go-junit-report > $(TEST_REPORT)

$(TEST_COVERAGE): $(TEST_REPORT)
	$(GO) tool cover -html=test-reports/coverprofile -o $(TEST_COVERAGE)

dist/www-growse-com_linux_%: $(TEST_COVERAGE)
	GOOS=linux GOARCH=$* $(GO) build -o dist/www-growse-com_linux_$*

.PHONY: clean
clean:
	rm -rf dist
	rm -rf test-reports
	rm -rf *.deb
