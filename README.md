# go-broadcaster
a rewrite of Sourcefabric's Airtime in Go for better performance and lower resource usage

## Structure

go-broadcaster will consist of 3 min Go programs.

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
  * implement basic functionality of playout-engine
  * implement playlist creation
