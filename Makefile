
#TOOLS

GO			=	go
GODEP		=	dep
GOEXE 	:= $(shell go env GOEXE)

NPM			= yarn
NPMRUN	= yarn

MODE ?= production

default: build

build: mediamonitor web

mediamonitor: _godeps
	$(GO) build -o ./bin/media-monitor$(GOEXE) ./cmd/media-monitor

web: webserver

webserver: _godeps webclient
ifneq ($(MODE), production)
	$(GO) build -o ./bin/webserver$(GOEXE) -tags=dev ./web
else
	$(GO)	generate ./web/client
	$(GO) build -o ./bin/webserver$(GOEXE) ./web
endif

webclient: _npmdeps
ifneq ($(MODE), production)
	cd ./web/client && $(NPMRUN) dev-build
else
	cd ./web/client && $(NPMRUN) build
endif
	mkdir -p ./bin/dist
	cp -r ./web/client/dist ./bin/dist

_godeps: go.mod
	$(GO) mod download

_npmdeps:
	$(NPM) install

tools: dbproto migrate rolemod usermod

dbproto: ./tools/dbproto/*
	mkdir -p ./bin/tools
	$(GO) build -o ./bin/tools/dbproto$(GOEXE) ./tools/dbproto

migrate: ./tools/migrate/*
	mkdir -p ./bin/tools/migrate
	$(GO) build -o ./bin/tools/migrate/migrate$(GOEXE) ./tools/migrate
	cp ./tools/migrate/*.sql ./bin/tools/migrate

rolemod: ./tools/rolemod/*
	mkdir -p ./bin/tools
	$(GO) build -o ./bin/tools/rolemod$(GOEXE) ./tools/rolemod

usermod: ./tools/usermod/*
	mkdir -p ./bin/tools
	$(GO) build -o ./bin/tools/usermod$(GOEXE) ./tools/usermod


clean:
	rm -rf ./bin
