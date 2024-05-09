define indent
  LEVEL=$(("$1 + 1"))
  INDENT="s/^/$(seq -s\  "$LEVEL" | tr -d '[:digit:]')/"
  case $(uname) in
	Darwin) sed -l "$INDENT";;
	*)      sed -u "$INDENT";;
  esac
endef
