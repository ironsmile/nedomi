# vim:noet
BINARY=nedomi
SOURCES := $(shell find . -name '*.go')

VERSION = $(shell cat VERSION)
BUILD_TIME = $(shell date +%s)
GIT_HASH ?= $(shell git show --pretty=%h -s HEAD)
GIT_TAG ?= $(shell git name-rev --tags --no-undefined --name-only HEAD 2>/dev/null)
GIT_STATUS ?= $(shell git status --porcelain -uno)
ifneq "$(GIT_STATUS)" ""
	DIRTY:=true
endif

LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitHash=${GIT_HASH} -X main.GitTag=${GIT_TAG} -X main.Dirty=${DIRTY}"

.DEFAULT_GOAL: $(BINARY)

default: ${BINARY}

${BINARY}: ${SOURCES}
	go build ${LDFLAGS} -o ${BINARY} main.go

.PHONY: install
install:
	go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
