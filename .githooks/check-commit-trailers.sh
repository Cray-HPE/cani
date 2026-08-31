#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2026 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#

# check-commit-trailers.sh validates the attribution trailers of a commit
# message: who co-authored the change, which tools assisted, and who signed
# it off. It is the local companion to check-conventional-commit.sh, which
# validates the header.
#
# Whether a change was AI-assisted cannot be detected, only declared, so this
# script checks that whatever was declared is well formed and machine
# readable. It also enforces the one rule an agent must never break: only a
# human can sign off a commit.
#
# It deliberately depends on nothing but a POSIX shell.

set -e
set -u

PROG=${0##*/}

# Attribution trailer tokens recognised by this repository, in canonical form.
ATTRIBUTION_TOKENS="Assisted-by Co-authored-by Signed-off-by"

# Name words that identify an AI assistant rather than a person. Kept to
# branded terms only; first names such as Claude or Gemini are ambiguous.
AI_NAME_WORDS="copilot chatgpt openai anthropic codex llm gpt"

# Mail domains belonging to AI vendors.
AI_MAIL_DOMAINS="openai.com anthropic.com"

# A literal carriage return, so messages written with CRLF endings validate
# the same as messages written with LF endings.
CR=$(printf '\r')

usage() {
  cat <<EOF
usage: $PROG --file <path> | --message <string>

  --file <path>       read the commit message from <path> (commit-msg hook)
  --message <string>  validate the given message string
  -h, --help          show this help
EOF
}

print_examples() {
  cat <<EOF
Attribution trailers go in the trailer block, one blank line after the body:

  feat(export): add rack filtering

  Assisted-by: GitHub Copilot
  Co-authored-by: Ada Lovelace <ada@example.com>
  Signed-off-by: Grace Hopper <grace@example.com>

Rules:
  - Assisted-by names the tool that helped, not an identity, so it carries
    no mail address
  - Co-authored-by credits a person or account and needs 'Name <user@host>'
  - Signed-off-by needs 'Name <user@host>' and must be a human; an AI cannot
    certify a contribution
  - the block is the last paragraph of the message and holds nothing but
    trailers; an attribution trailer anywhere else is rejected
  - tokens are spelled: $ATTRIBUTION_TOKENS

See .github/CONTRIBUTING.md for details.
EOF
}

fail() {
  printf 'commit message rejected: %s\n\n' "$1" >&2
  printf 'offending trailer:\n  %s\n\n' "$2" >&2
  print_examples >&2
  exit 1
}

lower() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]'
}

# name_words prints a value as lower-case, space delimited words so that an
# identity matches whole words only ("Hellman" must not match "llm").
name_words() {
  printf '%s' "$1" | tr -c '[:alnum:]' ' ' | tr '[:upper:]' '[:lower:]'
}

# trailer_token prints the token of a git trailer line, or nothing when the
# line is not shaped like a trailer.
trailer_token() {
  token=""

  case "$1" in
    *:*) ;;
    *) return 0 ;;
  esac

  token=${1%%:*}
  case "$token" in
    '' | *[![:alnum:]-]*) return 0 ;;
  esac

  printf '%s' "$token"
}

# canonical_attribution prints the canonical spelling of an attribution
# token, or nothing when the token is some other trailer such as 'Refs'.
canonical_attribution() {
  needle=$(lower "$1")

  for known in $ATTRIBUTION_TOKENS; do
    if [ "$needle" = "$(lower "$known")" ]; then
      printf '%s' "$known"
      return 0
    fi
  done

  return 0
}

# trim prints a value without leading or trailing whitespace. Tabs count as
# whitespace, so a trailer whose value is only a tab reads as empty.
trim() {
  trimmed=$1

  while :; do
    case "$trimmed" in
      [[:space:]]*) trimmed=${trimmed#?} ;;
      *) break ;;
    esac
  done
  while :; do
    case "$trimmed" in
      *[[:space:]]) trimmed=${trimmed%?} ;;
      *) break ;;
    esac
  done

  printf '%s' "$trimmed"
}

# trailer_value prints the value of a trailer line.
trailer_value() {
  trim "${1#*:}"
}

# has_identity succeeds when a value is a 'Name <user@host.tld>' pair. The
# address is split into its parts rather than pattern matched as a whole, so
# that an empty user, host or domain label is rejected.
has_identity() {
  name=""
  address=""
  user=""
  host=""
  tld=""

  case "$1" in
    *" <"*">") ;;
    *) return 1 ;;
  esac

  name=${1%% <*}
  [ -n "$name" ] || return 1

  address=${1#*<}
  address=${address%>*}
  case "$address" in
    *@*@* | *@) return 1 ;;
    *@*) ;;
    *) return 1 ;;
  esac

  user=${address%@*}
  case "$user" in
    '' | *[[:space:]]* | *[\<\>]*) return 1 ;;
  esac

  host=${address#*@}
  case "$host" in
    .* | *. | *..* | *[![:alnum:].-]*) return 1 ;;
    *.*) ;;
    *) return 1 ;;
  esac

  tld=${host##*.}
  case "$tld" in
    ? | *[![:alpha:]]*) return 1 ;;
  esac

  return 0
}

# is_ai_mail_domain succeeds when a value carries an AI vendor mail address.
is_ai_mail_domain() {
  domain=${1#*@}
  domain=${domain%%>*}
  domain=$(lower "$domain")

  for known in $AI_MAIL_DOMAINS; do
    if [ "$domain" = "$known" ]; then
      return 0
    fi
  done

  return 1
}

# is_ai_identity succeeds when a value names an AI assistant.
is_ai_identity() {
  haystack=" $(name_words "$1") "

  for word in $AI_NAME_WORDS; do
    case "$haystack" in
      *" $word "*) return 0 ;;
    esac
  done

  is_ai_mail_domain "$1"
}

check_assisted_by() {
  if [ -z "$1" ]; then
    fail "Assisted-by must name the tool that assisted" "$2"
  fi

  case "$1" in
    *"<"*"@"*) fail "Assisted-by names a tool, not an identity; drop the mail address" "$2" ;;
  esac
}

check_co_authored_by() {
  if ! has_identity "$1"; then
    fail "Co-authored-by must be 'Name <user@host>'" "$2"
  fi

  if is_ai_mail_domain "$1"; then
    fail "Co-authored-by credits a person or account; declare tools with 'Assisted-by:'" "$2"
  fi
}

check_signed_off_by() {
  if ! has_identity "$1"; then
    fail "Signed-off-by must be 'Name <user@host>'" "$2"
  fi

  if is_ai_identity "$1"; then
    fail "only a human can sign off a commit; declare the tool with 'Assisted-by:'" "$2"
  fi
}

check_trailer() {
  value=$(trailer_value "$2")

  case "$1" in
    Assisted-by) check_assisted_by "$value" "$2" ;;
    Co-authored-by) check_co_authored_by "$value" "$2" ;;
    Signed-off-by) check_signed_off_by "$value" "$2" ;;
  esac
}

# is_continuation succeeds for an indented, non-blank line, which git treats
# as the folded remainder of the trailer above it.
is_continuation() {
  case "$1" in
    *[![:space:]]*) ;;
    *) return 1 ;;
  esac

  case "$1" in
    [[:space:]]*) return 0 ;;
  esac

  return 1
}

# clean_message prints a commit message without the parts git drops before it
# looks for trailers: comment lines, the --verbose diff after the scissors and
# carriage returns. Folded values are joined onto their trailer so that later
# passes see one whole value per line, the way git reports them.
clean_message() {
  line=""
  held=""
  holding=0

  while IFS= read -r line || [ -n "$line" ]; do
    line=${line%"$CR"}

    case "$line" in
      '# ------------------------ >8'*) break ;;
      '#'*) continue ;;
    esac

    if [ "$holding" -eq 1 ] && is_continuation "$line"; then
      held="$held $(trim "$line")"
      continue
    fi

    [ "$holding" -eq 0 ] || printf '%s\n' "$held"
    held=$line
    holding=1
  done

  [ "$holding" -eq 0 ] || printf '%s\n' "$held"
}

# line_kind classifies a message line as blank, trailer or text.
line_kind() {
  case "$1" in
    *[![:space:]]*) ;;
    *)
      printf 'blank'
      return 0
      ;;
  esac

  if [ -n "$(trailer_token "$1")" ]; then
    printf 'trailer'
  else
    printf 'text'
  fi
}

# trailer_block_start reads a message on stdin and prints the line number
# where the final trailer block begins, or 0 when there is none. The block is
# the last paragraph, preceded by a blank line, whose every line is a trailer.
# git also accepts a paragraph that is merely mostly trailers, so this is the
# stricter rule and never credits a trailer git would not see.
trailer_block_start() {
  lineno=0
  para_start=0
  para_ok=0
  last_start=0
  last_ok=0
  line=""
  kind=""

  while IFS= read -r line || [ -n "$line" ]; do
    lineno=$((lineno + 1))
    kind=$(line_kind "$line")

    if [ "$kind" = blank ]; then
      if [ "$para_start" -gt 0 ]; then
        last_start=$para_start
        last_ok=$para_ok
        para_start=0
      fi
      continue
    fi

    if [ "$para_start" -eq 0 ]; then
      para_start=$lineno
      if [ "$kind" = trailer ]; then
        para_ok=1
      else
        para_ok=0
      fi
      continue
    fi

    case "$kind" in
      trailer) ;;
      *) para_ok=0 ;;
    esac
  done

  if [ "$para_start" -gt 0 ]; then
    last_start=$para_start
    last_ok=$para_ok
  fi

  # A one paragraph message is all subject, so it has no trailer block.
  if [ "$last_ok" -eq 1 ] && [ "$last_start" -gt 1 ]; then
    printf '%s' "$last_start"
  else
    printf '0'
  fi
}

# validate_trailers reads a message on stdin and checks every attribution
# trailer it declares, rejecting any that falls outside the final trailer
# block given as $1.
validate_trailers() {
  block=$1
  lineno=0
  line=""
  token=""
  canonical=""

  while IFS= read -r line || [ -n "$line" ]; do
    lineno=$((lineno + 1))

    token=$(trailer_token "$line")
    [ -n "$token" ] || continue

    canonical=$(canonical_attribution "$token")
    [ -n "$canonical" ] || continue

    if [ "$block" -eq 0 ] || [ "$lineno" -lt "$block" ]; then
      fail "attribution trailers belong in the last paragraph, one blank line after the body, and that paragraph must hold nothing but trailers" "$line"
    fi

    if [ "$token" != "$canonical" ]; then
      fail "spell the trailer as '$canonical:'" "$line"
    fi

    check_trailer "$canonical" "$line"
  done

  return 0
}

main() {
  mode=""
  value=""
  message=""
  block=0

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --file)
        [ "$#" -ge 2 ] || { printf '%s: --file requires a value\n' "$PROG" >&2; exit 2; }
        mode="file"
        value=$2
        shift 2
        ;;
      --message)
        [ "$#" -ge 2 ] || { printf '%s: --message requires a value\n' "$PROG" >&2; exit 2; }
        mode="message"
        value=$2
        shift 2
        ;;
      -h | --help)
        usage
        exit 0
        ;;
      *)
        printf '%s: unknown argument: %s\n' "$PROG" "$1" >&2
        usage >&2
        exit 2
        ;;
    esac
  done

  case "$mode" in
    file)
      [ -f "$value" ] || { printf '%s: file not found: %s\n' "$PROG" "$value" >&2; exit 2; }
      message=$(clean_message < "$value")
      ;;
    message)
      message=$(printf '%s\n' "$value" | clean_message)
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac

  block=$(printf '%s\n' "$message" | trailer_block_start)
  printf '%s\n' "$message" | validate_trailers "$block"
}

main "$@"
