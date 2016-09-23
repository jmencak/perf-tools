#!/bin/sh
# Written so that it can work even with busybox's "ash shell".

ProgramName=${0##*/}

# Global variables #############################################################
RUN=${RUN:-${ProgramName%.sh}}
LOGGING_DELAY=${LOGGING_DELAY:-1000000}	# delay $LOGGING_DELAY microseconds before sending another log line
LOGGING_LINE_LENGTH=${LOGGING_LINE_LENGTH:-80}
sleep_bin=usleep

fail()
{
  echo $@ >&2
}

warn()
{
  fail "$ProgramName: $@"
}

die()
{
  local err=$1
  shift
  fail "$ProgramName: $@"
  exit $err
}

usage()
{
  cat <<EOF 1>&2
Usage: $ProgramName
EOF
}

define_sleep()
{
  usleep 0 || {
    # let's hope we have sleep and that it supports real numbers...
    sleep 0.000001 || die 1 "Don't have usleep nor sleep that support real numbers"

    # have sleep that suports real numbers, don't rely on 'bc' being available
    sleep_bin=sleep
    expr $LOGGING_DELAY + 0 >/dev/null 2>&1 || die "LOGGING_DELAY not an integer"
    LOGGING_DELAY=`printf "%06d" $LOGGING_DELAY | sed 's|^\(.*\)\(......\)$|\1.\2|'`
  }
}

main_logger()
{
  charset='[:alnum:] \t'
  while true
  do 
    log_string=`tr -cd "$charset" < /dev/urandom | head -c $LOGGING_LINE_LENGTH`
    logger "$log_string"
    $sleep_bin $LOGGING_DELAY
  done
}

define_sleep
case "$RUN" in
  logger)
    main_logger "$@"
  ;;

  *) die 1 "mode \`$RUN' not supported"
esac
