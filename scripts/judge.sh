#!/usr/bin/env bash
# truthsayer-judge — LLM judgment layer for truthsayer findings
# Usage: truthsayer scan --format json . | ./scripts/judge.sh [--precedents FILE]
#
# Reads JSON findings from stdin, checks against precedents,
# calls LLM for unprecedented findings, outputs verdicts.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PRECEDENTS_FILE="${1:-$SCRIPT_DIR/../state/precedents.json}"
VERDICTS_FILE="${VERDICTS_FILE:-/dev/stdout}"
MODEL="${TRUTHSAYER_JUDGE_MODEL:-claude-haiku}"

mkdir -p "$(dirname "$PRECEDENTS_FILE")"

# Initialize precedents if missing
if [[ ! -f "$PRECEDENTS_FILE" ]]; then
  echo '[]' > "$PRECEDENTS_FILE"
fi

# Read findings from stdin
FINDINGS="$(cat)"
FINDING_COUNT=$(echo "$FINDINGS" | jq 'if type == "array" then length else .findings // [] | length end')

if [[ "$FINDING_COUNT" == "0" ]]; then
  echo '{"verdict": "clean", "guilty": 0, "not_guilty": 0, "advisory": 0, "findings": []}' | jq .
  exit 0
fi

# Extract findings array
FINDINGS_ARRAY=$(echo "$FINDINGS" | jq 'if type == "array" then . else .findings // [] end')

# Check each finding against precedents
UNPRECEDENTED='[]'
PRECEDENTED_VERDICTS='[]'

while IFS= read -r finding; do
  rule=$(echo "$finding" | jq -r '.rule_id // .rule // ""')
  file=$(echo "$finding" | jq -r '.file // .path // ""')
  snippet=$(echo "$finding" | jq -r '.line_text // .snippet // ""')
  
  # Check for matching precedent
  match=$(jq -r --arg rule "$rule" --arg snippet "$snippet" '
    .[] | select(.rule == $rule) |
    select(.confidence >= 0.9) |
    .verdict
  ' "$PRECEDENTS_FILE" | head -1)
  
  if [[ -n "$match" ]]; then
    # Established precedent — apply cached verdict
    PRECEDENTED_VERDICTS=$(echo "$PRECEDENTED_VERDICTS" | jq --argjson f "$finding" --arg v "$match" '. + [$f + {verdict: $v, source: "precedent"}]')
    # Bump seen count
    jq --arg rule "$rule" '
      map(if .rule == $rule then .seen_count += 1 | .last_seen = (now | strftime("%Y-%m-%d")) else . end)
    ' "$PRECEDENTS_FILE" > "$PRECEDENTS_FILE.tmp" && mv "$PRECEDENTS_FILE.tmp" "$PRECEDENTS_FILE"
  else
    UNPRECEDENTED=$(echo "$UNPRECEDENTED" | jq --argjson f "$finding" '. + [$f]')
  fi
done < <(echo "$FINDINGS_ARRAY" | jq -c '.[]')

UNPRECEDENTED_COUNT=$(echo "$UNPRECEDENTED" | jq 'length')

if [[ "$UNPRECEDENTED_COUNT" -gt 0 ]]; then
  # Batch unprecedented findings for LLM judgment
  # Read source context for each finding
  CONTEXT=""
  while IFS= read -r finding; do
    file=$(echo "$finding" | jq -r '.file // .path // ""')
    line=$(echo "$finding" | jq -r '.line // .line_number // 1')
    rule=$(echo "$finding" | jq -r '.rule_id // .rule // ""')
    msg=$(echo "$finding" | jq -r '.message // .description // ""')
    
    start=$((line > 5 ? line - 5 : 1))
    end=$((line + 5))
    
    if [[ -f "$file" ]]; then
      # REASON: intentional fallback - if sed fails, we show "(could not read)" instead
      code=$(sed -n "${start},${end}p" "$file" 2>/dev/null || echo "(could not read)")
    else
      code="(file not accessible)"
    fi
    
    CONTEXT+="
---
Rule: $rule
File: $file:$line
Message: $msg
Code context:
\`\`\`
$code
\`\`\`
"
  done < <(echo "$UNPRECEDENTED" | jq -c '.[]')

  # Call LLM
  PROMPT="You are Truthsayer's judge — you review code findings and render verdicts.

For each finding below, examine the code context and decide:
- **guilty**: This is a real problem that should be fixed. Explain why.
- **not-guilty**: This is a false positive or intentionally correct code. Explain why.
- **advisory**: This is real but minor — should be fixed eventually, not urgently.

Also suggest a precedent pattern (a generalized description) so similar findings can be auto-judged next time.

Respond as a JSON array:
[{\"rule\": \"...\", \"file\": \"...\", \"verdict\": \"guilty|not-guilty|advisory\", \"reasoning\": \"...\", \"precedent_pattern\": \"...\"}]

Findings to judge:
$CONTEXT"

  JUDGMENT=$(echo "$PROMPT" | curl -s https://api.anthropic.com/v1/messages \
    -H "x-api-key: $ANTHROPIC_API_KEY" \
    -H "anthropic-version: 2023-06-01" \
    -H "content-type: application/json" \
    -d "$(jq -cn --arg p "$PROMPT" '{model: "claude-haiku-4-20250514", max_tokens: 4096, messages: [{role: "user", content: $p}]}')" \
    | jq -r '.content[0].text // empty')

  if [[ -z "$JUDGMENT" ]]; then
    echo "Error: LLM judgment failed" >&2
    exit 1
  fi

  # Parse judgment and create precedents
  # REASON: jq parse errors handled by loop termination - no JSON means no iterations
  echo "$JUDGMENT" | jq -c '.[]' 2>/dev/null | while IFS= read -r j; do
    rule=$(echo "$j" | jq -r '.rule')
    verdict=$(echo "$j" | jq -r '.verdict')
    reasoning=$(echo "$j" | jq -r '.reasoning')
    pattern=$(echo "$j" | jq -r '.precedent_pattern')
    
    # Add to precedents
    jq --arg rule "$rule" --arg verdict "$verdict" --arg reasoning "$reasoning" \
       --arg pattern "$pattern" \
      '. + [{
        rule: $rule,
        pattern: $pattern,
        verdict: $verdict,
        reasoning: $reasoning,
        confidence: 0.5,
        seen_count: 1,
        first_seen: (now | strftime("%Y-%m-%d")),
        last_seen: (now | strftime("%Y-%m-%d"))
      }]' "$PRECEDENTS_FILE" > "$PRECEDENTS_FILE.tmp" && mv "$PRECEDENTS_FILE.tmp" "$PRECEDENTS_FILE"
  done

  # Merge all verdicts
  # REASON: jq parse error handled by fallback to empty array
  LLM_VERDICTS=$(echo "$JUDGMENT" | jq '[.[] | {rule, file, verdict, reasoning, source: "judge"}]' 2>/dev/null || echo '[]')
else
  LLM_VERDICTS='[]'
fi

# Final summary
ALL_VERDICTS=$(jq -cn --argjson p "$PRECEDENTED_VERDICTS" --argjson l "$LLM_VERDICTS" '$p + $l')
GUILTY=$(echo "$ALL_VERDICTS" | jq '[.[] | select(.verdict == "guilty")] | length')
NOT_GUILTY=$(echo "$ALL_VERDICTS" | jq '[.[] | select(.verdict == "not-guilty")] | length')
ADVISORY=$(echo "$ALL_VERDICTS" | jq '[.[] | select(.verdict == "advisory")] | length')

jq -cn \
  --arg v "$([ "$GUILTY" -gt 0 ] && echo "guilty" || echo "clean")" \
  --argjson g "$GUILTY" \
  --argjson ng "$NOT_GUILTY" \
  --argjson a "$ADVISORY" \
  --argjson findings "$ALL_VERDICTS" \
  '{verdict: $v, guilty: $g, not_guilty: $ng, advisory: $a, findings: $findings}'

# Exit 1 if any guilty verdicts (for use as gate)
if [[ "$GUILTY" -gt 0 ]]; then
  exit 1
fi
