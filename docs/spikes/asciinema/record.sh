#!/usr/bin/env bash
# Simulated-typing driver for the asciinema demos.
# Usage: asciinema rec --window-size 155x28 -c "bash record.sh status" demo-status.cast
set -u
export PATH="$HOME/.work/bin:$PATH"

type_run() {
  local cmd="$1"
  printf '\033[38;5;135m>\033[0m '
  local i
  for ((i = 0; i < ${#cmd}; i++)); do
    printf '%s' "${cmd:i:1}"
    sleep 0.015
  done
  sleep 0.4
  printf '\n'
  eval "$cmd"
  sleep 1.6
}

RULE_FLAGS="--match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region maximize \\
  --match 'app=\"Google Chrome\" title=/neozenith/'          --display 2 --region left-half"

case "${1:?demo name required: status|apply|profile}" in
status)
  type_run "screenz doctor"
  type_run "screenz status --match 'bundle=com.microsoft.VSCode title=/office-demo/' --match 'app=\"Google Chrome\" title=/neozenith/'"
  ;;
apply)
  type_run "screenz apply --dry-run \\
  $RULE_FLAGS"
  # exit code is echoed in the same eval so \$? is really apply's status
  type_run "screenz apply \\
  $RULE_FLAGS; echo exit=\$?"
  ;;
profile)
  SCREENZ_HOME="$(mktemp -d)"
  export SCREENZ_HOME
  type_run "screenz profile save office \\
  $RULE_FLAGS"
  type_run 'cat $SCREENZ_HOME/profiles/office.yaml'
  type_run "screenz apply --dry-run office"
  ;;
esac
sleep 0.8
