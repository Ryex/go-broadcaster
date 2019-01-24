#!/bin/sh

set -e

/migrate init || true
exec /migrate
