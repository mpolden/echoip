#!/usr/bin/env bash

NAME="ifconfig"
IFC_PATH=$(dirname $(readlink -f $0))
LOCK_FILE="$IFC_PATH/$NAME.lock"
PID_FILE="$IFC_PATH/$NAME.pid"
BIN="$IFC_PATH/$NAME"

E_USAGE=1
E_NOTFOUND=2
E_NOPID=3
E_LOCKED=4

if [[ ! -x "$BIN" ]]; then
    echo "$BIN does not exist or is not executable"
    exit $E_NOTFOUND
fi

case "$1" in
    start)
        if [[ -f "$LOCK_FILE" ]]; then
            echo "Lock file $LOCK_FILE exists. Already running?"
            exit $E_LOCKED
        fi
        echo "Starting $NAME"
        daemonize -c $IFC_PATH -o ${NAME}.log -p $PID_FILE -l $LOCK_FILE $BIN
        ;;
    stop)
        if [[ ! -s "$PID_FILE" ]]; then
            echo "PID file $PID_FILE empty or not found. Not started?"
            exit $E_NOPID
        fi
        PID=$(head -n1 $PID_FILE)
        echo "Stopping $NAME: $PID"
        kill $PID && rm -- $PID_FILE $LOCK_FILE
        ;;
    status)
        if [[ ! -s "$PID_FILE" ]]; then
            echo "$NAME is not running"
        else
            echo "$NAME is running: $(head -n1 $PID_FILE)"
        fi
        ;;
    *)
        echo "usage: $0 {start|stop|status}"
        exit $E_USAGE
        ;;
esac
