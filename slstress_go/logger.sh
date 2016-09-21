#!/bin/sh

ProgramName=${0##*/}

# Global variables #############################################################
RUN=${RUN:-${ProgramName%.sh}}
LOGGING_DELAY=${LOGGING_DELAY:-1000000}	# delay $LOGGING_DELAY microseconds before sending another log line
LOGGING_LINE_LENGTH=${LOGGING_LINE_LENGTH:-80}

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

main_logger()
{
  charset='[:alnum:] \t'
  while true
  do 
    log_string=`tr -cd "$charset" < /dev/urandom | head -c ${LOGGING_LINE_LENGTH}`
    logger ${log_string}
    usleep ${LOGGING_DELAY}
  done
}

case "$RUN" in
  logger)
    main_logger "$@"
  ;;

  *) die 1 "mode \`$RUN' not supported"
esac
