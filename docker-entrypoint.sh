#!/bin/bash
set -e

CMD=()
XP_UI_CONFIG_PATH=
XP_UI_DIST_DIR=${XP_UI_DIST_DIR:-}
XP_UI_DIST_CONFIG_FILE=${XP_UI_DIST_DIR}/app.config.js
XP_API_BIN=${XP_API_BIN:?"ERROR: XP_API_BIN is not specified"}

show_help() {
  cat <<EOF
Usage: $(basename "$0") <options> <...>
    -ui-config               JSON file containing configuration of XP UI
EOF
}

main(){
  parse_command_line "$@"

  if [[ -n "$XP_UI_CONFIG_PATH" ]]; then
    echo "XP UI config found at ${XP_UI_CONFIG_PATH}..."
    if [[ -n "$XP_UI_DIST_DIR" ]]; then
      echo "Overriding UI config at $XP_UI_DIST_CONFIG_FILE"

      echo "var xpConfig = $(cat $XP_UI_CONFIG_PATH);" > "$XP_UI_DIST_CONFIG_FILE"

      echo "Done."
    else
      echo "XP_UI_DIST_DIR: XP UI static build directory not provided. Skipping."
    fi
  else
    echo "XP UI config is not provided. Skipping."
  fi
}

parse_command_line(){
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -ui-config)
        if [[ -n "$2" ]]; then
          XP_UI_CONFIG_PATH="$2"
          shift
        else
          echo "ERROR: '-ui-config' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      *)
        CMD+=("$1")
        ;;
    esac

    shift
  done

  if [[ -n "$XP_UI_CONFIG_PATH" ]]; then
    if [ ! -f "$XP_UI_CONFIG_PATH" ]; then
      echo "ERROR: config file $XP_UI_CONFIG_PATH does not exist." >&2
      show_help
      exit 1
    fi
  fi
}

main "$@"

echo "Launching xp-management server: " "$XP_API_BIN" "${CMD[@]}"
exec "$XP_API_BIN" "${CMD[@]}"
