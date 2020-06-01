ARCH := amd64
VERSION := 1.0.3

BUILD_NUMBER ?= 0
DEBVERSION := $(VERSION:v%=%)-$(BUILD_NUMBER)
PKGNAME := growse-com-locator-and-search
TEST_COVERAGE := coverage.txt

LDFLAGS := "-w -s"

export GOPATH := $(shell go env GOPATH)

.PHONY: package

package: $(addsuffix .deb, $(addprefix $(PKGNAME)_$(VERSION)-$(BUILD_NUMBER)_, $(foreach a, $(ARCH), $(a))))

$(PKGNAME)_$(VERSION)-$(BUILD_NUMBER)_%.deb: dist/www-growse-com_linux_%
	chmod +x $<
	bundle exec fpm -f -s dir -t deb --url https://www.growse.com/ --description "growse.com dynamic content (locator, search)" --deb-systemd www-growse-com.service -n $(PKGNAME) --config-files /etc/www-growse-com.conf.json -p . -a $* -v $(DEBVERSION) $<=/usr/bin/www.growse.com www-growse-com.conf.json=/etc/www-growse-com.conf.json databasemigrations/=/var/www/growse-web/databasemigrations

.PHONY: test
test: $(TEST_COVERAGE)

$(TEST_COVERAGE):
	go test -cover -covermode=count -coverprofile=$@ -v

.PHONY: build
build: $(addprefix dist/www-growse-com_linux_, $(foreach a, $(ARCH), $(a)))

dist/www-growse-com_linux_%:
	go mod vendor -v
	GOOS=linux GOARCH=$* go build -ldflags=$(LDFLAGS) -o dist/www-growse-com_linux_$*
	upx $@

.PHONY: clean
clean:
	rm -rf dist
	rm -f $(TEST_COVERAGE)
	rm -rf *.deb
