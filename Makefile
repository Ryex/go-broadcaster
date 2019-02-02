
#TOOLS

GO			=	go
GODEP		=	dep

NPM			= yarn
NPMRUN	= yarn

DEVELOPMENT ?= no

PACKR = $(GOPATH)/bin/packr

default: build

build: mediamonitor web

mediamonitor: _godeps
	$(GO) build -o ./bin/media-monitor ./cmd/media-monitor

web: webserver

webserver: _godeps webclient
ifneq ($(DEVELOPMENT), yes)
	$(PACKR)
endif
	$(GO) build -o ./bin/webserver ./web/
ifneq ($(DEVELOPMENT), yes)
	$(PACKR) clean
endif
	cp -r ./web/dist ./bin

webclient: _npmdeps
ifeq ($(DEVELOPMENT), yes)
	 cd ./web && $(NPMRUN) devbuild
else
	 cd ./web && $(NPMRUN) build
endif



_godeps:
	$(GODEP) ensure

_npmdeps:
	$(NPM) install
