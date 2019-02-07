
#TOOLS

GO			=	go
GODEP		=	dep
GOEXE 	:= $(shell go env GOEXE)

NPM			= yarn
NPMRUN	= yarn

MODE ?= production

default: build

build: mediamon webserver

mediamon: _godeps
	$(GO) build -o ./bin/gobroadcaster-mediamon$(GOEXE) ./cmd/gobroadcaster-mediamon

webserver: _godeps webclient
	mkdir -p ./bin
ifneq ($(MODE), production)
	$(GO) build -o ./bin/gobroadcaster-web$(GOEXE) -tags=dev ./cmd/gobroadcaster-web
else
	cp -r ./bin/dist ./cmd.gobroadcaster-web/client
	$(GO)	generate ./cmd/gobroadcaster-web/client
	rm -rf ./cmd.gobroadcaster-web/client/dist
	$(GO) build -o ./bin/gobroadcaster-web$(GOEXE) ./cmd/gobroadcaster-web
endif

webclient: _npmdeps
	mkdir -p ./bin
ifneq ($(MODE), production)
	cd ./web && $(NPMRUN) dev-build
else
	cd ./web && $(NPMRUN) build
endif
	cp -r ./web/dist ./bin/dist

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
