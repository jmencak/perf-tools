#!/bin/sh
# Written so that it can work even with busybox's "ash shell".

ProgramName=${0##*/}

# Global variables
url_gun_ws="http://${GUN}:9090"

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

# Wait for all the pods to be in the Running state
synchronize_pods()
{
  while [[ -z $(curl -s ${url_gun_ws}) ]]
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
  local err="$1"

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

define_timeout_bin

case "${RUN}" in
  stress)
    synchronize_pods
    
    [[ "${STRESS_CPU}" ]] && STRESS_CPU="--cpu ${STRESS_CPU}"
#    [[ "${STRESS_TIME}" ]] && STRESS_TIME="--timeout ${STRESS_TIME}"
    $timeout \
      stress ${STRESS_CPU}
  ;;
  slstress)
    synchronize_pods

    $timeout \
      /usr/local/bin/slstress.sh "$@" > *.log

#    if [[ "${pbench_dir}" == *pbench-user-benchmark* ]]; then
#      # Copy results back to Cluster Loader host in PBench dir
#      scp *.jtl *.log *.png ${GUN}:${pbench_dir}
#    fi
  ;;
  logger)
    synchronize_pods

    $timeout \
      /usr/local/bin/slstress.sh "$@"

#    if [[ "${pbench_dir}" == *pbench-user-benchmark* ]]; then
#      # Copy results back to Cluster Loader host in PBench dir
#      scp *.jtl *.log *.png ${GUN}:${pbench_dir}
#    fi
  ;;
  jmeter)
    die 1 "${RUN} not supported."
  ;; 
  *)
    die 1 "Need to specify what to run."
  ;;
esac
timeout_exit_status
