REPO ?= quay.io/faroshq/
TAG_NAME ?= $(shell git describe --tags --abbrev=0)
KO_DOCKER_REPO ?= ${REPO}

lint:
	gofmt -s -w cmd hack pkg
	go run golang.org/x/tools/cmd/goimports -w -local=github.com/kube-red cmd hack pkg
	go run ./hack/validate-imports cmd hack pkg
	staticcheck ./...

setup-kind:
	./hack/dev/setup-kind.sh

delete-kind:
	./hack/dev/delete-kind.sh
	rm -rf dev/database.sqlite3

images:
	KO_DOCKER_REPO=${KO_DOCKER_REPO} ko build --sbom=none -B --platform=linux/amd64 -t latest ./cmd/*

show-sqlite-database:
	sqlitebrowser dev/database.sqlite3

dev-run-hello-world:
	docker run -it -p 8080:8080 quay.io/synpse/hello-synpse-go

