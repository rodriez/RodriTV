#!/usr/bin/env bash
# Build component_index.md from Java and Go source code.
# Requires only POSIX tools available on Linux and macOS: bash, awk, find, sort.
#
# Usage:
#   ./build_component_index.sh <project_root> [--output component_index.md]
#
# Scans all .java and .go files under <project_root>, extracts modules
# (packages), components (classes / interfaces / enums / records / structs /
# top-level functions) and methods with their signatures, and writes a
# Markdown tree to component_index.md.
set -u

ROOT=""
OUT="component_index.md"
while [ $# -gt 0 ]; do
  case "$1" in
    --output) OUT="$2"; shift 2 ;;
    *) ROOT="$1"; shift ;;
  esac
done
if [ -z "$ROOT" ] || [ ! -d "$ROOT" ]; then
  echo "usage: $0 <project_root> [--output component_index.md]" >&2
  exit 1
fi
ROOT=$(cd "$ROOT" && pwd)

RECORDS=$(mktemp "${TMPDIR:-/tmp}/compidx.XXXXXX")
trap 'rm -f "$RECORDS"' EXIT

# ---------------------------------------------------------------------------
# Java parser (awk). Emits TSV records:
#   module \t folder \t comp \t file \t kind \t seq \t method \t signature
# Component records have empty method/signature; method records follow them.
# ---------------------------------------------------------------------------
JAVA_AWK='
function strip_comments(  i, c, nxt, out) {
  out = ""
  for (i = 1; i <= length($0); i++) {
    c = substr($0, i, 1); nxt = substr($0, i + 1, 1)
    if (inblock) { if (c == "*" && nxt == "/") { inblock = 0; i++ }; continue }
    if (c == "/" && nxt == "*") { inblock = 1; i++; continue }
    if (c == "/" && nxt == "/") break
    out = out c
  }
  $0 = out
}
function countch(s, ch,   i, n) { n = 0; for (i = 1; i <= length(s); i++) if (substr(s, i, 1) == ch) n++; return n }
function emit(kind, meth, sig) {
  printf "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n", module, folder, comp, fname, kind, ++seq, meth, sig >> RECS
}
{
  strip_comments()
  line = $0
  sub(/^[ \t]+/, "", line); sub(/[ \t]+$/, "", line)

  if (module == "" && line ~ /^package[ \t]+[A-Za-z0-9_.]+/) {
    module = line; sub(/^package[ \t]+/, "", module); sub(/;.*/, "", module)
  }

  opens = countch(line, "{"); closes = countch(line, "}")

  # ---- type declaration -------------------------------------------------
  if (!intype && line ~ /(class|interface|enum|record)[ \t]+[A-Za-z_]/ && line ~ /\{/) {
    s = line
    while (s ~ /^(public|protected|private|abstract|final|static|sealed)[ \t]+/) sub(/^[a-z]+[ \t]+/, "", s)
    kind = ""; comp = ""
    if (match(s, /@?interface[ \t]+[A-Za-z_][A-Za-z0-9_]*/) || match(s, /(class|enum|record)[ \t]+[A-Za-z_][A-Za-z0-9_]*/)) {
      kw = substr(s, RSTART, RLENGTH)
      split(kw, kwp, /[ \t]+/); kind = kwp[1]; comp = kwp[2]
      sub(/^@/, "", kind)
    }
    if (comp != "") {
      intype = 1; typedepth = depth + opens
      emit((kind == "class" ? "class" : kind), "", "")
      depth += opens - closes
      next
    }
  }

  # ---- method candidate at member depth ---------------------------------
  if (intype && depth == typedepth && line ~ /\(/) {
    buf = line
    # accumulate until parens balance
    while (countch(buf, "(") > countch(buf, ")") && (getline nl) > 0) {
      $0 = nl; strip_comments(); nl2 = $0
      sub(/^[ \t]+/, "", nl2); buf = buf " " nl2
    }
    s = buf
    # strip leading annotations/modifiers only; keep annotations inside params
    sub(/^[ \t]+/, "", s)
    while (s ~ /^@[A-Za-z0-9_.]+(\([^()]*\))?[ \t]*/) sub(/^@[A-Za-z0-9_.]+(\([^()]*\))?[ \t]*/, "", s)
    while (s ~ /^(public|private|protected|static|final|abstract|synchronized|native|default|strictfp)[ \t]+/) sub(/^[a-z]+[ \t]+/, "", s)
    if (substr(s, 1, 1) == "<") { gp = index(s, ">"); if (gp > 0) { s = substr(s, gp + 1); sub(/^[ \t]+/, "", s) } }
    p = index(s, "(")
    if (p > 1) {
      head = substr(s, 1, p - 1); sub(/[ \t]+$/, "", head)
      d2 = 0; q = 0
      for (i = p; i <= length(s); i++) { c = substr(s, i, 1); if (c == "(") d2++; else if (c == ")") { d2--; if (d2 == 0) { q = i; break } } }
      if (q > 0) {
        params = substr(s, p + 1, q - p - 1)
        rest = substr(s, q + 1)
        r = rest; gsub(/throws[ \t][A-Za-z0-9_., \t]+/, "", r)
        if (r ~ /^[ \t]*[{;]/ && head !~ /=/) {
          ntok = split(head, tok, /[ \t]+/)
          mname = tok[ntok]
          ret = substr(head, 1, length(head) - length(mname)); sub(/[ \t]+$/, "", ret)
          firsttok = ret; sub(/[ \t].*/, "", firsttok)
          if (mname !~ /^(if|for|while|switch|catch|return|new|do|else|try|synchronized|throw)$/ && \
              firsttok !~ /^(return|new|throw|if|for|while|switch|case|break|continue)$/) {
            gsub(/[ \t]+/, " ", params)
            sub(/^[ \t]+/, "", params)
            if (mname == comp) sig = mname "(" params ")"
            else sig = mname "(" params ") : " ret
            emit("", mname, sig)
          }
        }
      }
    }
  }

  depth += opens - closes
  if (intype && depth < typedepth) intype = 0
}
'

# ---------------------------------------------------------------------------
# Go parser (awk). Same TSV record format.
# ---------------------------------------------------------------------------
GO_AWK='
function strip_comments(  i, c, nxt, out) {
  out = ""
  for (i = 1; i <= length($0); i++) {
    c = substr($0, i, 1); nxt = substr($0, i + 1, 1)
    if (inblock) { if (c == "*" && nxt == "/") { inblock = 0; i++ }; continue }
    if (c == "/" && nxt == "*") { inblock = 1; i++; continue }
    if (c == "/" && nxt == "/") break
    out = out c
  }
  $0 = out
}
function countch(s, ch,   i, n) { n = 0; for (i = 1; i <= length(s); i++) if (substr(s, i, 1) == ch) n++; return n }
function emit(comp, kind, meth, sig) {
  m = (module == "" ? "(unknown)" : module)
  printf "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n", m, folder, comp, fname, kind, ++seq, meth, sig >> RECS
}
{
  strip_comments()
  line = $0
  sub(/^[ \t]+/, "", line); sub(/[ \t]+$/, "", line)

  if (module == "" && line ~ /^package[ \t]+[A-Za-z_]/) {
    module = line; sub(/^package[ \t]+/, "", module); sub(/[ \t].*/, "", module)
  }

  # ---- struct / interface types ------------------------------------------
  if (line ~ /^type[ \t]+[A-Za-z_][A-Za-z0-9_]*[ \t]+(struct|interface)[ \t]*\{/) {
    comp = line; sub(/^type[ \t]+/, "", comp); sub(/[ \t].*/, "", comp)
    kind = (line ~ /struct/) ? "struct" : "interface"
    emit(comp, kind, "", "")
    next
  }

  # ---- functions / methods ------------------------------------------------
  if (line ~ /^func[ \t]/) {
    buf = line
    while (countch(buf, "(") > countch(buf, ")") && (getline nl) > 0) {
      nl2 = nl; sub(/^[ \t]+/, "", nl2); buf = buf " " nl2
    }
    s = buf; sub(/^func[ \t]+/, "", s)
    recv = ""
    if (substr(s, 1, 1) == "(") {
      q = index(s, ")")
      if (q == 0) next
      recv = substr(s, 2, q - 2)
      s = substr(s, q + 1); sub(/^[ \t]+/, "", s)
      nr = split(recv, rt, /[ \t]+/); recv = rt[nr]; gsub(/\*/, "", recv)    }
    p = index(s, "(")
    if (p == 0) next
    mname = substr(s, 1, p - 1); sub(/[ \t]+$/, "", mname)
    sub(/\[[^]]*\]/, "", mname)   # drop generic type params
    d2 = 0; q = 0
    for (i = p; i <= length(s); i++) { c = substr(s, i, 1); if (c == "(") d2++; else if (c == ")") { d2--; if (d2 == 0) { q = i; break } } }
    if (q == 0) next
    params = substr(s, p + 1, q - p - 1)
    ret = substr(s, q + 1); sub(/\{.*/, "", ret)
    sub(/^[ \t]+/, "", ret); sub(/[ \t]+$/, "", ret)
    gsub(/[ \t]+/, " ", params)
    sig = mname "(" params ")" (ret != "" ? " " ret : "")
    if (recv != "") emit(recv, "", mname, sig)
    else emit(mname, "func", "", "")
  }
}
'

# ---------------------------------------------------------------------------
# Renderer (awk): TSV records -> Markdown tree.
# ---------------------------------------------------------------------------
RENDER_AWK='
BEGIN { FS = "\t" }
{ rec[NR] = $0 }
END {
  print "# Component Index"; print ""
  prevkey = ""; prevmod = ""; prevfolder = ""
  for (i = 1; i <= NR; i++) {
    split(rec[i], f, FS)
    mod = f[1]; folder = f[2]; comp = f[3]; file = f[4]; kind = f[5]; meth = f[7]; sig = f[8]
    ck = mod SUBSEP folder
    if (ck != prevkey) { print mod " (" folder ")"; prevmod = mod; prevfolder = folder; prevkey = ck; prevcomp = "" }
    compkey = comp SUBSEP file
    if (meth == "") {
      # component line: is it the last component of this folder?
      last = 1
      for (j = i + 1; j <= NR; j++) {
        split(rec[j], g, FS)
        if (g[1] SUBSEP g[2] != ck) break
        if (g[3] SUBSEP g[4] != compkey && g[7] == "") { last = 0; break }
      }
      tag = (kind == "struct" || kind == "interface" || kind == "func" || kind == "enum" || kind == "record" || kind == "type") ? " [" kind "]" : ""
      printf "%s %s%s (%s)\n", (last ? "└──" : "├──"), comp, tag, file
      curindent = (last ? "    " : "│   ")
      prevcomp = compkey
    } else {
      # method line: last method of its component?
      last = (i == NR)
      if (!last) { split(rec[i + 1], g, FS); if (g[3] SUBSEP g[4] != compkey || g[7] == "") last = 1 }
      printf "%s%s %s (%s)\n", curindent, (last ? "└──" : "├──"), meth, sig
    }
    # blank line when folder ends
    if (i == NR) { print "" }
    else { split(rec[i + 1], g, FS); if (g[1] SUBSEP g[2] != ck) print "" }
  }
}
'

n_java=0
n_go=0
seq=0

while IFS= read -r f; do
  rel=${f#"$ROOT"/}
  folder=${rel%/*}
  [ "$folder" = "$rel" ] && folder="/"
  fname=${rel##*/}
  case "$f" in
    *.java)
      awk -v RECS="$RECORDS" -v module="" -v folder="$folder" -v fname="$fname" -v seq="$seq" "$JAVA_AWK" "$f"
      seq=$(wc -l < "$RECORDS" | tr -d ' ')
      n_java=$((n_java + 1)) ;;
    *.go)
      awk -v RECS="$RECORDS" -v module="" -v folder="$folder" -v fname="$fname" -v seq="$seq" "$GO_AWK" "$f"
      seq=$(wc -l < "$RECORDS" | tr -d ' ')
      n_go=$((n_go + 1)) ;;
  esac
done <<EOF
$(find "$ROOT" \( -name .git -o -name node_modules -o -name vendor -o -name target -o -name build -o -name dist -o -name out -o -name .idea -o -name .gradle -o -name bin -o -name testdata \) -prune -o -type f \( -name '*.java' -o -name '*.go' \) -print | sort)
EOF

if [ ! -s "$RECORDS" ]; then
  echo "warn: no components found under $ROOT" >&2
fi

LC_ALL=C sort -t "$(printf '\t')" -k1,1 -k2,2 -k3,3 -k4,4 -k6,6n "$RECORDS" | awk "$RENDER_AWK" > "$OUT"

echo "Scanned $((n_java + n_go)) files (Java: $n_java, Go: $n_go)."
echo "Wrote $OUT"
