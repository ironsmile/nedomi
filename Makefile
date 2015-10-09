# vim:noet
BINARY=nedomi
SOURCES := $(shell find . -name '*.go')

VERSION=$(shell cat VERSION)
BUILD_TIME=$(shell date +%FT%T%z)
GIT_HASH=$(shell git show --pretty=%h -s HEAD)
GIT_TAG=$(shell git name-rev --tags --no-undefined --name-only HEAD 2>/dev/null)
GIT_STATUS := $(shell git status --porcelain -uno)

ifneq "$(GIT_STATUS)" ""
	GIT_HASH:=${GIT_HASH}-dirty
ifneq "$(GIT_TAG)" ""
	GIT_TAG:=${GIT_TAG}-dirty
endif
endif

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitHash=${GIT_HASH} -X main.GitTag=${GIT_TAG}"

.DEFAULT_GOAL: $(BINARY)

nedomi: ${SOURCES}
	go build ${LDFLAGS} -o nedomi main.go

.PHONY: install
install:
	go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
