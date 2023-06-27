#! /usr/bin/env bash

kubectl get pod/$POD_NAME -n $NAMESPACE 1>&2
if [ $? -ne 0 ]; then
  echo "[#] starting pod/$POD_NAME in namespace '$NAMESPACE'"
  kubectl run $POD_NAME --namespace $NAMESPACE --image=nxtcoder17/alpine.python3:nonroot --restart=Never -- sh -c 'tail -f /dev/null' 2>&1 /dev/null
  sleep 5
else
  echo "pod/$POD_NAME is already running in namespace '$NAMESPACE'"
fi
