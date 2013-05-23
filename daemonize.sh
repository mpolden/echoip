#!/usr/bin/env bash

NAME="ifconfig"
PREFIX=$(dirname $(readlink -f $0))
DAEMON="$PREFIX/$NAME"
PID_FILE="${PREFIX}/tmp/${NAME}.pid"
LOCK_FILE="${PREFIX}/tmp/${NAME}.lock"
LOG_FILE="${PREFIX}/tmp/${NAME}.log"

E_USAGE=1
E_NOTFOUND=2

if [[ ! -x "$DAEMON" ]]; then
    echo "$DAEMON does not exist or is not executable"
    exit $E_NOTFOUND
fi

start () {
    echo -n "Starting $NAME: "
    mkdir -p $PREFIX/tmp
    daemonize -c $PREFIX -o $LOG_FILE -p $PID_FILE -l $LOCK_FILE $DAEMON && \
        echo "ok" || echo "failed"
}

stop () {
    echo -n "Stopping $NAME: "
    if [[ -s "$PID_FILE" ]]; then
        PID=$(head -n1 $PID_FILE)
        kill $PID 2> /dev/null && echo "ok" || echo "not running?"
    fi
    rm -f -- $PID_FILE $LOCK_FILE
}

status () {
    if [[ -s "$PID_FILE" ]]; then
        PID=$(head -n1 $PID_FILE)
        kill -0 $PID 2> /dev/null && echo "$NAME is running (pid: $PID)" || \
            echo "$NAME is not running"
    else
        echo "$NAME is not running"
    fi
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        start
        ;;
    status)
        status
        ;;
    *)
        echo "usage: $0 {start|stop|restart|status}"
        exit $E_USAGE
        ;;
esac

exit $?
