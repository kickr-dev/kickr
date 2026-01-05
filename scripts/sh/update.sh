#!/bin/sh

log_info() {
  fg="\033[0;34m"
  reset="\033[0m"
  echo "${fg}$1${reset}"
}

cmd=""
for c in kickr.dev kickr; do
  [ "$cmd" = "" ] || command -v $c > /dev/null 2>&1 || continue
  cmd=$c
done
if [ "$cmd" = "" ]; then
  echo "No kickr generator found, exiting"
  exit 2
fi
log_info "Found kickr generator named '$cmd'"

workspaces=$(find / -name workspaces 2>/dev/null)
for workspace in $workspaces; do
  dirs=$(find "$workspace" -name testdata -prune -o -name .kickr -exec dirname {} +;)
  for dir in $dirs; do
    log_info "Updating layout of $dir"
    $cmd --dir "$dir" generate
  done
  unset dirs dir
done
unset workspaces workspace
