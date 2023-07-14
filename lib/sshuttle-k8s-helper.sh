#! /usr/bin/env bash

set -o nounset
set -o pipefail
set -o errexit

tFile=$(mktemp)
debug=false

[ $debug = true ] && echo "$@" > "$tFile"
podRef=$1
IFS=/; read -a items <<< "$podRef"

if [ ${#items} -gt 1 ]; then
  namespace=${items[0]}
  podName=${items[1]}
else
  podName=${items[0]}
fi
[ $debug = true ] && echo "$1" >> "$tFile"
shift 2;
[ $debug = true ] && echo "$@" >> "$tFile"

if [ "$namespace" != "" ]; then
  [ $debug = true ] && echo "exec kubectl exec -n $namespace -i $podName -- $*" >> "$tFile"
  kubectl exec -n "$namespace" -i "$podName" -- "$@"
else
  [ $debug = true ] && echo "exec kubectl exec -i $podName -- $*" >> "$tFile"
  kubectl exec -i "$podName" -- "$@"
fi
