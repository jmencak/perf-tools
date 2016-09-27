#!/bin/sh
# Written so that it can work even with busybox's "ash shell".
# Entrypoint script for Load Generator Docker Image

ProgramName=${0##*/}

# Global variables
url_gun_ws="http://${GUN}:9090"
#gateway=$(/sbin/ip route|awk '/default/ { print $3 }')	# sometimes there is no /sbin/ip ...
gw_hex=$(grep ^eth0 /proc/net/route | head -1 | awk '{print $3}')
gateway=$(printf "%d.%d.%d.%d" 0x${gw_hex:6:2} 0x${gw_hex:4:2} 0x${gw_hex:2:2} 0x${gw_hex:0:2})

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

have_server()
{
  local server="$1"
  if test "${server}" = "127.0.0.1" || test "${server}" = "" ; then
    # server not defined
    return 1
  fi 
}

have_gun()
{
  have_server "${GUN}"
}

have_pbench()
{
  have_server "${PBENCH_HOST}"
}

# Wait for all the pods to be in the Running state
synchronize_pods()
{
  have_gun || return

  while [ -z $(curl -s "${url_gun_ws}") ]
  do 
    sleep 5
    fail "${url_gun_ws} not ready"
  done
}

# basic checks for toybox/busybox/coreutils timeout
define_timeout_bin()
{
  test "${RUN_TIMEOUT}" || return	# timeout empty, do not define it and just return

  timeout -t 0 /bin/sleep 0 >/dev/null 2>&1

  case $? in
    0)   # we have a busybox timeout with '-t' option for number of seconds
       timeout="timeout -t ${RUN_TIMEOUT}"
    ;;
    1)   # we have toybox's timeout without the '-t' option for number of seconds
       timeout="timeout ${RUN_TIMEOUT}"
    ;;
    125) # we have coreutil's timeout without the '-t' option for number of seconds
       timeout="timeout ${RUN_TIMEOUT}"
    ;;
    *)   # couldn't find timeout or unknown version
       warn "running without timeout"
       timeout=""
    ;;
  esac
}

timeout_exit_status()
{
  local err="${1:-$?}"

  case $err in
    124) # coreutil's return code for timeout
       return 0
    ;;
    143) # busybox's return code for timeout with default signal TERM
       return 0
    ;;
    *) return $err
  esac
}

main()
{
  define_timeout_bin

  case "${RUN}" in
    stress)
      synchronize_pods
 
      [ "${STRESS_CPU}" ] && STRESS_CPU="--cpu ${STRESS_CPU}"
      $timeout \
        stress ${STRESS_CPU}
      ;;

    slstress)
      local slstress_log=/tmp/${HOSTNAME}-${gateway}.log

      synchronize_pods
      $timeout \
        slstress \
          -l ${LOGGING_LINE_LENGTH} \
          -w \
          ${LOGGING_DELAY} > ${slstress_log}
      $(timeout_exit_status) || exit $?	# slstress failed, exit

      if have_pbench ; then
        scp -p ${slstress_log} ${PBENCH_HOST}:${PBENCH_DIR}
      fi
    ;;

    logger)
      local slstress_log=/tmp/${HOSTNAME}-${gateway}.log

      synchronize_pods
      $timeout \
        /usr/local/bin/logger.sh
      $(timeout_exit_status) || exit $?	# logger failed, exit
    ;;

    jmeter)
      die 1 "${RUN} not supported."
    ;; 

    *)
      die 1 "Need to specify what to run."
    ;;
  esac
  timeout_exit_status
}

main
