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

workspaces=$(find "$HOME" -type d -name workspaces -prune -print 2>/dev/null)
[ ! -d /workspaces ] || workspaces="$workspaces /workspaces"
for workspace in $workspaces; do
  dirs=$(find "$workspace" \
    \( -name .git -o -name .gitlab-ci-local -o -name .terraform -o -name node_modules -o -name testdata -o -name vendor \) -prune \
    -o -type f -name .kickr.yml -exec dirname {} +)
  for dir in $dirs; do
    (
      $cmd --dir "$dir" --log-level warn
      [ -z "$(git -C "$dir" status --porcelain -- . ':(exclude,glob)**/.gitignore')" ] || log_info "Updated layout of $dir"
    ) &
  done
  unset dirs dir
done
wait
unset workspaces workspace
