#!/usr/bin/env bash
# envx — quick tour.
#
# This script drives the asciinema recording embedded in the README. Run it
# standalone with envx on your PATH:
#
#   bash docs/demo.sh
#
# To regenerate the recording and GIF:
#
#   go build -o /tmp/envxrec/envx ./cmd/envx
#   asciinema rec --overwrite --window-size 92x24 -f asciicast-v2 \
#     -c "env ENVX_BIN_DIR=/tmp/envxrec bash docs/demo.sh" docs/demo.cast
#   agg docs/demo.cast docs/demo.gif
set -e

# Allow the recorder to point at a freshly built binary; otherwise use PATH.
[ -n "$ENVX_BIN_DIR" ] && export PATH="$ENVX_BIN_DIR:$PATH"
unset NO_COLOR
cd "$(mktemp -d)"

CYAN='\033[1;36m'
BOLD='\033[1m'
RESET='\033[0m'

# run echoes a command at a shell-like prompt, then executes it.
run() {
	printf "${CYAN}\$${RESET} ${BOLD}%s${RESET}\n" "$*"
	sleep 0.7
	"$@"
	echo
	sleep 1.1
}

sleep 0.6
run envx init
run envx profile add dev
run envx set PORT 8080 --profile dev
run envx set API_KEY s3cr3t-token --profile dev --sensitive
run envx show dev
run envx profile add prod --extends dev
run envx set PORT 443 --profile prod
run envx diff dev prod
run envx export --profile prod --format dotenv
sleep 1.5
