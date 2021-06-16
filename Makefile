OS = $(shell uname | tr A-Z a-z)

REVIVE_VERSION = 1.0.7
STATICCHECK_VERSION = 2021.1

bin/revive: bin/revive-${REVIVE_VERSION}
	@ln -sf revive-${REVIVE_VERSION} bin/revive
bin/revive-${REVIVE_VERSION}:
	@mkdir -p bin
	curl -L https://github.com/mgechev/revive/releases/download/v${REVIVE_VERSION}/revive_${REVIVE_VERSION}_$(shell uname)_x86_64.tar.gz | tar -zOxf - revive > ./bin/revive-${REVIVE_VERSION} && chmod +x ./bin/revive-${REVIVE_VERSION}

bin/staticcheck: bin/staticcheck-${STATICCHECK_VERSION}
	@ln -sf staticcheck-${STATICCHECK_VERSION} bin/staticcheck
bin/staticcheck-${STATICCHECK_VERSION}:
	@mkdir -p bin
	curl -L https://github.com/dominikh/go-tools/releases/download/${STATICCHECK_VERSION}/staticcheck_${OS}_amd64.tar.gz | tar -zOxf - staticcheck > ./bin/staticcheck-${STATICCHECK_VERSION} && chmod +x ./bin/staticcheck-${STATICCHECK_VERSION}

.PHONY: lint
lint: bin/revive bin/staticcheck
	go vet ./...
	bin/revive ./...
	bin/staticcheck ./...
	gofmt -l -s -e . | grep .go && exit 1
