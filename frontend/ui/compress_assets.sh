#/usr/bin/env bash
#
# compress static assets

set -eu

cp embed.go.tmpl embed.go

GZIP_OPTS="-fk"
# gzip option '-k' may not always exist in the latest gzip available on different distros
if ! gzip -k -h &>/dev/null; then GZIP_OPTS="-f"; fi

find build -type f -name '*.gz' -delete
find build -type f -exec gzip $GZIP_OPTS '{}' \; -print0 | xargs -0 -I % echo %.gz | xargs echo //go:embed >> embed.go
echo var EmbedFS embed.FS >> embed.go
