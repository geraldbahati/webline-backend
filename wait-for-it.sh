#!/usr/bin/env bash
# Use this script to test if a given TCP host/port are available

WAITFORIT_cmdname=${0##*/}

# Colors for better output readability
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to show error messages
echoerr() {
    if [[ $WAITFORIT_QUIET -ne 1 ]]; then
        echo -e "${RED}[ERROR]${NC} $*" 1>&2
    fi
}

# Function to show info messages
echoinfo() {
    if [[ $WAITFORIT_QUIET -ne 1 ]]; then
        echo -e "${BLUE}[INFO]${NC} $*"
    fi
}

# Function to show success messages
echosuccess() {
    if [[ $WAITFORIT_QUIET -ne 1 ]]; then
        echo -e "${GREEN}[SUCCESS]${NC} $*"
    fi
}

# Function to show warning messages
echowarn() {
    if [[ $WAITFORIT_QUIET -ne 1 ]]; then
        echo -e "${YELLOW}[WARNING]${NC} $*"
    fi
}

usage()
{
    cat << USAGE >&2
Usage:
    $WAITFORIT_cmdname host:port [-s] [-t timeout] [-- command args]
    -h HOST | --host=HOST       Host or IP under test
    -p PORT | --port=PORT       TCP port under test
                                Alternatively, you specify the host and port as host:port
    -s | --strict               Only execute subcommand if the test succeeds
    -q | --quiet                Don't output any status messages
    -t TIMEOUT | --timeout=TIMEOUT
                                Timeout in seconds, zero for no timeout
    -r RETRIES | --retries=RETRIES
                                Number of retries before giving up (default: 1)
    -i INTERVAL | --interval=INTERVAL
                                Seconds between retries (default: 1)
    -- COMMAND ARGS             Execute command with args after the test finishes
USAGE
    exit 1
}

wait_for()
{
    if [[ $WAITFORIT_TIMEOUT -gt 0 ]]; then
        echoinfo "Waiting $WAITFORIT_TIMEOUT seconds for $WAITFORIT_HOST:$WAITFORIT_PORT"
    else
        echoinfo "Waiting for $WAITFORIT_HOST:$WAITFORIT_PORT without a timeout"
    fi

    WAITFORIT_start_ts=$(date +%s)
    retries=0
    max_retries=${WAITFORIT_RETRIES:-1}
    wait_interval=${WAITFORIT_INTERVAL:-1}

    while [[ $retries -lt $max_retries ]]; do
        if [[ $WAITFORIT_ISBUSY -eq 1 ]]; then
            nc -z "$WAITFORIT_HOST" "$WAITFORIT_PORT"
            WAITFORIT_result=$?
        else
            (echo -n > /dev/tcp/$WAITFORIT_HOST/$WAITFORIT_PORT) >/dev/null 2>&1
            WAITFORIT_result=$?
        fi

        if [[ $WAITFORIT_result -eq 0 ]]; then
            WAITFORIT_end_ts=$(date +%s)
            seconds=$((WAITFORIT_end_ts - WAITFORIT_start_ts))
            echosuccess "$WAITFORIT_HOST:$WAITFORIT_PORT is available after $seconds seconds"
            break
        fi

        retries=$((retries+1))

        # If we've reached max retries or would exceed the timeout
        if [[ $retries -ge $max_retries ]]; then
            echoerr "Failed to connect to $WAITFORIT_HOST:$WAITFORIT_PORT after $max_retries attempts"
            break
        fi

        # Calculate remaining time if timeout is specified
        if [[ $WAITFORIT_TIMEOUT -gt 0 ]]; then
            current_ts=$(date +%s)
            elapsed=$((current_ts - WAITFORIT_start_ts))
            remaining=$((WAITFORIT_TIMEOUT - elapsed))

            if [[ $remaining -le 0 ]]; then
                echoerr "Timeout occurred after waiting $elapsed seconds for $WAITFORIT_HOST:$WAITFORIT_PORT"
                break
            fi

            echowarn "Connection attempt $retries failed. Retrying in $wait_interval seconds ($remaining seconds remaining)..."
        else
            echowarn "Connection attempt $retries failed. Retrying in $wait_interval seconds..."
        fi

        sleep $wait_interval
    done

    return $WAITFORIT_result
}

wait_for_wrapper()
{
    # In order to support SIGINT during timeout: http://unix.stackexchange.com/a/57692
    if [[ $WAITFORIT_QUIET -eq 1 ]]; then
        timeout $WAITFORIT_BUSYTIMEFLAG $WAITFORIT_TIMEOUT $0 --quiet --child --host=$WAITFORIT_HOST --port=$WAITFORIT_PORT --timeout=$WAITFORIT_TIMEOUT --retries=$WAITFORIT_RETRIES --interval=$WAITFORIT_INTERVAL &
    else
        timeout $WAITFORIT_BUSYTIMEFLAG $WAITFORIT_TIMEOUT $0 --child --host=$WAITFORIT_HOST --port=$WAITFORIT_PORT --timeout=$WAITFORIT_TIMEOUT --retries=$WAITFORIT_RETRIES --interval=$WAITFORIT_INTERVAL &
    fi
    WAITFORIT_PID=$!
    trap "kill -INT -$WAITFORIT_PID" INT
    wait $WAITFORIT_PID
    WAITFORIT_RESULT=$?
    if [[ $WAITFORIT_RESULT -ne 0 ]]; then
        echoerr "Timeout occurred after waiting $WAITFORIT_TIMEOUT seconds for $WAITFORIT_HOST:$WAITFORIT_PORT"
    fi
    return $WAITFORIT_RESULT
}

# Process arguments
WAITFORIT_HOST=""
WAITFORIT_PORT=""
WAITFORIT_TIMEOUT=15
WAITFORIT_STRICT=0
WAITFORIT_CHILD=0
WAITFORIT_QUIET=0
WAITFORIT_RETRIES=3
WAITFORIT_INTERVAL=1

while [[ $# -gt 0 ]]
do
    case "$1" in
        *:* )
        WAITFORIT_hostport=(${1//:/ })
        WAITFORIT_HOST=${WAITFORIT_hostport[0]}
        WAITFORIT_PORT=${WAITFORIT_hostport[1]}
        shift 1
        ;;
        --child)
        WAITFORIT_CHILD=1
        shift 1
        ;;
        -q | --quiet)
        WAITFORIT_QUIET=1
        shift 1
        ;;
        -s | --strict)
        WAITFORIT_STRICT=1
        shift 1
        ;;
        -h)
        WAITFORIT_HOST="$2"
        if [[ $WAITFORIT_HOST == "" ]]; then break; fi
        shift 2
        ;;
        --host=*)
        WAITFORIT_HOST="${1#*=}"
        shift 1
        ;;
        -p)
        WAITFORIT_PORT="$2"
        if [[ $WAITFORIT_PORT == "" ]]; then break; fi
        shift 2
        ;;
        --port=*)
        WAITFORIT_PORT="${1#*=}"
        shift 1
        ;;
        -t)
        WAITFORIT_TIMEOUT="$2"
        if [[ $WAITFORIT_TIMEOUT == "" ]]; then break; fi
        shift 2
        ;;
        --timeout=*)
        WAITFORIT_TIMEOUT="${1#*=}"
        shift 1
        ;;
        -r)
        WAITFORIT_RETRIES="$2"
        if [[ $WAITFORIT_RETRIES == "" ]]; then break; fi
        shift 2
        ;;
        --retries=*)
        WAITFORIT_RETRIES="${1#*=}"
        shift 1
        ;;
        -i)
        WAITFORIT_INTERVAL="$2"
        if [[ $WAITFORIT_INTERVAL == "" ]]; then break; fi
        shift 2
        ;;
        --interval=*)
        WAITFORIT_INTERVAL="${1#*=}"
        shift 1
        ;;
        --)
        shift
        WAITFORIT_CLI=("$@")
        break
        ;;
        --help)
        usage
        ;;
        *)
        echoerr "Unknown argument: $1"
        usage
        ;;
    esac
done

if [[ "$WAITFORIT_HOST" == "" || "$WAITFORIT_PORT" == "" ]]; then
    echoerr "Error: you need to provide a host and port to test."
    usage
fi

WAITFORIT_TIMEOUT=${WAITFORIT_TIMEOUT:-15}
WAITFORIT_STRICT=${WAITFORIT_STRICT:-0}
WAITFORIT_CHILD=${WAITFORIT_CHILD:-0}
WAITFORIT_QUIET=${WAITFORIT_QUIET:-0}
WAITFORIT_RETRIES=${WAITFORIT_RETRIES:-3}
WAITFORIT_INTERVAL=${WAITFORIT_INTERVAL:-1}

# Check to see if timeout is from busybox?
WAITFORIT_TIMEOUT_PATH=$(type -p timeout)
WAITFORIT_TIMEOUT_PATH=$(realpath $WAITFORIT_TIMEOUT_PATH 2>/dev/null || readlink -f $WAITFORIT_TIMEOUT_PATH)

WAITFORIT_BUSYTIMEFLAG=""
if [[ $WAITFORIT_TIMEOUT_PATH =~ "busybox" ]]; then
    WAITFORIT_ISBUSY=1
    # Check if busybox timeout uses -t flag
    # (recent Alpine versions don't support -t anymore)
    if timeout &>/dev/stdout | grep -q -e '-t '; then
        WAITFORIT_BUSYTIMEFLAG="-t"
    fi
else
    WAITFORIT_ISBUSY=0
fi

# Validate the port number
if ! [[ "$WAITFORIT_PORT" =~ ^[0-9]+$ ]]; then
    echoerr "Error: Port must be a number"
    exit 1
fi

# Validate timeout is a number
if ! [[ "$WAITFORIT_TIMEOUT" =~ ^[0-9]+$ ]]; then
    echoerr "Error: Timeout must be a number"
    exit 1
fi

# Validate retries is a number
if ! [[ "$WAITFORIT_RETRIES" =~ ^[0-9]+$ ]]; then
    echoerr "Error: Retries must be a number"
    exit 1
fi

# Validate interval is a number
if ! [[ "$WAITFORIT_INTERVAL" =~ ^[0-9]+$ ]]; then
    echoerr "Error: Interval must be a number"
    exit 1
fi

if [[ $WAITFORIT_CHILD -gt 0 ]]; then
    wait_for
    WAITFORIT_RESULT=$?
    exit $WAITFORIT_RESULT
else
    if [[ $WAITFORIT_TIMEOUT -gt 0 ]]; then
        wait_for_wrapper
        WAITFORIT_RESULT=$?
    else
        wait_for
        WAITFORIT_RESULT=$?
    fi
fi

if [[ $WAITFORIT_RESULT -ne 0 && $WAITFORIT_STRICT -eq 1 ]]; then
    echoerr "$WAITFORIT_cmdname: strict mode, refusing to execute subprocess"
    exit $WAITFORIT_RESULT
fi

if [[ $WAITFORIT_CLI != "" ]]; then
    if [[ $WAITFORIT_RESULT -ne 0 && $WAITFORIT_STRICT -eq 1 ]]; then
        echoerr "$WAITFORIT_cmdname: strict mode, refusing to execute subprocess"
        exit $WAITFORIT_RESULT
    fi
    exec "${WAITFORIT_CLI[@]}"
else
    exit $WAITFORIT_RESULT
fi
