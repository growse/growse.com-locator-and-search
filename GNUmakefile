ARCH := amd64
VERSION := 1.0.3

BUILD_NUMBER ?= 0
DEBVERSION := $(VERSION:v%=%)-$(BUILD_NUMBER)
PKGNAME := growse-com-locator-and-search
TEST_REPORT := test-reports/report.xml
TEST_COVERAGE := test-reports/coverage.html

export GOPATH := $(shell go env GOPATH)

.PHONY: package

package: $(addsuffix .deb, $(addprefix $(PKGNAME)_$(VERSION)-$(BUILD_NUMBER)_, $(foreach a, $(ARCH), $(a))))

$(PKGNAME)_$(VERSION)-$(BUILD_NUMBER)_%.deb: dist/www-growse-com_linux_%
	chmod +x $<
	bundle exec fpm -f -s dir -t deb --url https://www.growse.com/ --description "growse.com dynamic content (locator, search)" --deb-systemd www-growse-com.service -n $(PKGNAME) --config-files /etc/www-growse-com.conf -p . -a $* -v $(DEBVERSION) $<=/usr/bin/www.growse.com config.json=/etc/www-growse-com.conf databasemigrations/=/var/www/growse-web/databasemigrations templates/=/var/www/growse-web/templates

$(GOPATH)/bin/go-junit-report:
	go get -u github.com/jstemmer/go-junit-report

.PHONY: test
test: $(TEST_COVERAGE)

$(TEST_REPORT): $(GOPATH)/bin/go-junit-report
	mkdir -p test-reports
	go test -cover -covermode=count -coverprofile=test-reports/coverprofile -v | tee /dev/tty | $(GOPATH)/bin/go-junit-report > $(TEST_REPORT)

$(TEST_COVERAGE): $(TEST_REPORT)
	go tool cover -html=test-reports/coverprofile -o $(TEST_COVERAGE)

.PHONY: build
build: $(addprefix dist/www-growse-com_linux_, $(foreach a, $(ARCH), $(a)))

dist/www-growse-com_linux_%:
	go mod vendor -v
	GOOS=linux GOARCH=$* go build -o dist/www-growse-com_linux_$*

.PHONY: clean
clean:
	rm -rf dist
	rm -rf test-reports
	rm -rf *.deb
