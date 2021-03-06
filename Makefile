.PHONY: get-tools test lint unit

get-tools:
	go get github.com/go-playground/overalls
	go get github.com/mattn/goveralls

test: lint unit

lint:
	go fmt ./...
	go vet ./...

unit:
	go test ./...

coverage:
	overalls -project=github.com/privacylab/talek -covermode=count -debug
	goveralls -coverprofile=overalls.coverprofile -service=travis-ci

ci: SHELL:=/bin/bash   # HERE: this is setting the shell for 'ci' only
ci: lint
	overalls -project=github.com/privacylab/talek -covermode=count -debug -- -tags travis
	@if [[ "${TRAVIS_JOB_NUMBER}" =~ ".1" ]]; then\
		echo "Uploading coverage to Coveralls.io"; \
		goveralls -coverprofile=overalls.coverprofile -service=travis-ci; \
	fi

docker-build:
	docker build -t talek-base:latest ./
	docker build -t talek-replica:latest ./cli/talekreplica/

docker-bash:
	docker run -it talek-base:latest bash
