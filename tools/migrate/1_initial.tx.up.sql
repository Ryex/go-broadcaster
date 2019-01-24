CREATE TABLE "library_paths" (
  "id" bigserial,
  "path" text,
  "added" timestamptz DEFAULT now(),
  "last_index" timestamptz,
  "indexing" boolean,
  PRIMARY KEY ("id")
);
CREATE TABLE "tracks" (
  "id" bigserial,
  "title" text,
  "album" text,
  "artist" text,
  "genre" text,
  "year" bigint,
  "length" bigint,
  "bitrate" bigint,
  "channels" bigint,
  "samplerate" bigint,
  "path" text,
  "added" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);
CREATE TABLE "users" (
  "id" bigserial,
  "username" text UNIQUE,
  "password" text,
  "roles" jsonb,
  "created_at" timestamptz DEFAULT now(),
  PRIMARY KEY ("id")
);
CREATE TABLE "roles" (
  "id" bigserial,
  "id_str" text UNIQUE,
  PRIMARY KEY ("id")
);
