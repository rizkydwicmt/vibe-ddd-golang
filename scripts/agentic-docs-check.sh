#!/usr/bin/env bash
set -euo pipefail

root=$(git rev-parse --show-toplevel)
cd "$root"

failed=0
skill_name_file=$(mktemp)
trap 'rm -f "$skill_name_file"' EXIT

if [[ ! -L AGENTS.md || "$(readlink AGENTS.md)" != "CLAUDE.md" ]]; then
  printf 'AGENTS.md must symlink directly to CLAUDE.md\n' >&2
  failed=1
fi

for source in .claude/skills/*/SKILL.md; do
  name=$(basename "$(dirname "$source")")
  mirror=".agents/skills/$name/SKILL.md"
  expected="../../../$source"
  if [[ ! -L "$mirror" || "$(readlink "$mirror")" != "$expected" ]]; then
    printf 'Missing or incorrect Codex skill mirror: %s -> %s\n' "$mirror" "$source" >&2
    failed=1
  fi
done

for mirror in .agents/skills/*/SKILL.md; do
  name=$(basename "$(dirname "$mirror")")
  if [[ ! -f ".claude/skills/$name/SKILL.md" ]]; then
    printf 'Stale Codex skill mirror without Claude source: %s\n' "$mirror" >&2
    failed=1
  fi
done

check_registered() {
  local kind=$1
  local path=$2
  local display_path=${path#./}

  if ! grep -Fq "($display_path)" CLAUDE.md; then
    printf 'Missing CLAUDE.md registration: %s (%s)\n' "$kind" "$display_path" >&2
    failed=1
  fi
}

while IFS= read -r command; do
  name="/${command##*/}"
  name=${name%.md}
  check_registered "$name" "$command"

  if ! grep -Eq '\.claude/skills/[^` )]+/SKILL\.md' "$command"; then
    printf 'Command has no authoritative skill reference: %s\n' "$command" >&2
    failed=1
  fi
done < <(find .claude/commands -type f -name '*.md' | sort)

while IFS= read -r skill; do
  name=$(sed -n 's/^name: //p' "$skill" | head -1)
  if [[ -z "$name" ]]; then
    printf 'Missing skill frontmatter name: %s\n' "$skill" >&2
    failed=1
    continue
  fi
  printf '%s\t%s\n' "$name" "$skill" >> "$skill_name_file"
  check_registered "$name" "$skill"
done < <(find .claude/skills -type f -name 'SKILL.md' | sort)

while IFS= read -r duplicate; do
  printf 'Duplicate skill name: %s\n' "$duplicate" >&2
  failed=1
done < <(cut -f1 "$skill_name_file" | sort | uniq -d)

while IFS= read -r agent; do
  name=$(sed -n 's/^name: //p' "$agent" | head -1)
  if [[ -z "$name" ]]; then
    printf 'Missing agent frontmatter name: %s\n' "$agent" >&2
    failed=1
    continue
  fi
  check_registered "$name" "$agent"

  if ! grep -Eq '\.claude/skills/[^` )]+/SKILL\.md' "$agent"; then
    printf 'Agent has no authoritative skill reference: %s\n' "$agent" >&2
    failed=1
  fi
done < <(find .claude/agents -type f -name '*.md' | sort)

while IFS= read -r reference; do
  if [[ ! -f "$reference" ]]; then
    printf 'Broken skill reference: %s\n' "$reference" >&2
    failed=1
  fi
done < <(grep -rhoE '\.claude/skills/[^` )]+/SKILL\.md' .claude/commands .claude/agents | sort -u)

if find .claude/commands .claude/agents -type f -name '*.md' -print0 \
  | xargs -0 grep -nE '^## (Hard [Rr]ules|Common [Mm]istakes|Preconditions|Acceptance|Environment [Rr]ules)$'; then
  printf 'Policy section found outside a skill; move it to the authoritative SKILL.md\n' >&2
  failed=1
fi

check_numbered_headers() {
  local directory=$1
  local file base number expected header expected_prefix

  while IFS= read -r file; do
    base=$(basename "$file")
    number=${base%%-*}
    expected=$number
    if [[ "$number" == "0000" ]]; then
      expected=NNNN
    fi

    header=$(sed -n '1p' "$file")
    expected_prefix="# ${expected}. "
    if [[ $header != "$expected_prefix"* ]]; then
      printf 'Numbered doc heading must match filename: %s (expected prefix: %s)\n' \
        "$file" "$expected_prefix" >&2
      failed=1
    fi
  done < <(find "$directory" -maxdepth 1 -type f -name '[0-9][0-9][0-9][0-9]-*.md' | sort)
}

check_numbered_headers docs/adr
check_numbered_headers docs/diagram

# ponytail: static checks cover template placeholders, not full Mermaid grammar; add
# mermaid-cli when arbitrary hand-written diagrams become common enough to justify it.
while IFS= read -r file; do
  while IFS= read -r violation; do
    [[ -z "$violation" ]] && continue
    printf 'Render-unsafe placeholder inside Mermaid block: %s:%s\n' "$file" "$violation" >&2
    failed=1
  done < <(awk '
    /^```mermaid[[:space:]]*$/ { in_mermaid = 1; next }
    /^```[[:space:]]*$/ { in_mermaid = 0; next }
    in_mermaid && /<[[:alnum:]][^>]*>/ { print FNR ":" $0 }
  ' "$file")
done < <(find docs -type f -name '*.md' | sort)

exit "$failed"
