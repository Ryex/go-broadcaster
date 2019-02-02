
#TOOLS

GO			=	go
GODEP		=	dep

NPM			= yarn
NPMRUN	= yarn

DEVELOPMENT ?= no

default: build

build: mediamonitor web

mediamonitor: _godeps
	$(GO) build -o ./bin/media-monitor ./cmd/media-monitor

web: webserver

webserver: _godeps webclient
ifneq ($(DEVELOPMENT), yes)
	$(GO) build -o ./bin/webserver -tags=dev ./web
else
	$(GO)	generate ./web/filesystem
	$(GO) build -o ./bin/webserver ./web
endif

webclient: _npmdeps
ifeq ($(DEVELOPMENT), yes)
	 cd ./web/client && $(NPMRUN) devbuild
else
	 cd ./web/client && $(NPMRUN) build
endif

_godeps: go.mod
	$(GO) mod download

_npmdeps:
	$(NPM) install
