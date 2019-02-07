# go-broadcaster
A rewrite of Sourcefabric's Airtime in Go for better performance and lower resource usage

## Structure

go-broadcaster will consist of 3 main Go programs.

* A web frontend (currently go-broadcaster-web)
* A media library monitor (currently media-monitor)
* A playout engine (not currently structured)

these three parts will communicate with a PostgresDB database and synchronize using
Postgres `NOTIFY` and `LISTEN`.

This *should* result in a cleaner and more maintainable system than the current PHP Zend implementation. The use of Go should make for a much faster and more responsive system, and using `NOTIFY/LISTEN` will eliminate the need for a rabbitmq server making the system much easier to scale.

This will let the web frontend schedule content and the playout engine will be notified of changes to parts of the schedule it may of already pulled.

The playout engine will this use a telnet connection to a liquidsoap server to control playback of content.

Initial design will focus around making playlists and scheduling them for playback.

## TODO
  * implement monitoring of, and importing of media in, library paths stored in the database
    - switch to tagLib wrapper
  * implement basic functionality of playout-engine
  * implement playlist creation

## Build
The Project provides a Makefile.
it is known to be built successfully ATM in the following environments:

  * Linux x86_64, Go 1.11.4+, yarn 1.13.0+, node v8.11.1+, GNU Make 4.2.1
  * Windows 10 /w MSYS2, Go 1.11.4+, yarn 1.13.0+, node v8.11.1+, GNU Make 4.2.1

in general running `make` in the project repository will build the entire project and place the output in `./bin`.

using `-e MODE=production` or `-e MODE=dev` with `make` will explicitly set the build mode.

### Build Examples

  ```
  ./go-broadcaster$ make
  ```

  ```
  ./go-broadcaster$ make -e MODE=production
  ```

  ```
  ./go-broadcaster$ make -e MODE=dev
  ```

### The hard way

The project consists of four main parts; the three go cmds, and the vue.js SPA web client.


#### Build the SPA client

The first step is to build the web client SPA

  1) cd to `./web` and run `yarn install` to install all the dependencies
  2) run `yarn build` or `yarn dev-build` to build out the desired files to `./web/dist`

This will get the SPA built to ./web/dist.

Next build the gobcasst-web binary. This can either embed the `./web/dist` SPA OR run
run with it off the disk from the working directory

#### Building Embedded

  1) copy `./web/dist` to `./cmd/gobcast-web/client/dist`

    cp -r ./web/dist ./cmd/gobcast-web/client

  2) run `go generate ./cmd/gobcast-web/client` to generate the embedded asset code
  3) `go build ./cmd/gobcast-web`

#### Building non Embedded

  1) `go build -tags dev ./cmd/gobcast-web`
  2) ensure the `./web/dist` directory is copied to the working directory when `gobcast-web` is run

#### Build the media monitor

  1) `go build ./cmd/gobcast-mediamon`


### TOOLS

  There are some useful tools located in `./tools`
  the most important of which is probably `migrate`

#### migrate
  Used to prep the database for the application

#### dbproto
  Used to prototype the database from the defined model structs.
  Mostly useful during development to quickly drop an create schema.

#### rolemod
  Used to add, remove, modify, and inspect roles in the database
  useful for development and potentially maintenance

#### usermod
Used to add, remove, modify, and inspect users in the database
useful for development and potentially maintenance
