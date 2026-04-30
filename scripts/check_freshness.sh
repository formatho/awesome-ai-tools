#!/bin/bash
# Freshness Checker for Awesome AI Tools
# Updates last_commit_date for all GitHub repos

# Get script directory (works in both CI and local environments)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
TOOLS_JSON="$REPO_DIR/data/tools.json"
TEMP_JSON="$REPO_DIR/data/tools_temp.json"

echo "🔍 Checking freshness of GitHub repos..."

# Backup current file
cp "$TOOLS_JSON" "$TOOLS_JSON.backup"

# Process each tool with GitHub URL
cat "$TOOLS_JSON" | jq -c '.tools[] | select(.url | contains("github.com"))' | while read -r tool; do
    REPO_URL=$(echo "$tool" | jq -r '.url')
    TOOL_NAME=$(echo "$tool" | jq -r '.name')
    
    # Convert GitHub URL to API URL
    API_URL=$(echo "$REPO_URL" | sed 's|https://github.com/|https://api.github.com/repos/|')
    
    # Get last commit date
    echo "Checking: $TOOL_NAME"
    LAST_COMMIT=$(curl -s "$API_URL" | jq -r '.pushed_at // empty')
    
    if [ -n "$LAST_COMMIT" ] && [ "$LAST_COMMIT" != "null" ]; then
        # Calculate days since last commit
        COMMIT_DATE=$(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$LAST_COMMIT" "+%s" 2>/dev/null || date -d "$LAST_COMMIT" "+%s" 2>/dev/null)
        NOW=$(date "+%s")
        DAYS=$(( (NOW - COMMIT_DATE) / 86400 ))
        
        # Determine freshness category
        if [ $DAYS -le 7 ]; then
            FRESHNESS="🟢 Fresh"
        elif [ $DAYS -le 30 ]; then
            FRESHNESS="🟡 Recent"
        elif [ $DAYS -le 90 ]; then
            FRESHNESS="🟠 Aging"
        else
            FRESHNESS="🔴 Stale"
        fi
        
        # Update JSON
        jq --arg name "$TOOL_NAME" \
           --arg date "$LAST_COMMIT" \
           --arg days "$DAYS" \
           --arg freshness "$FRESHNESS" \
           '(.tools[] | select(.name == $name)) |= . + {last_commit_date: $date, days_since_commit: ($days | tonumber), freshness: $freshness}' \
           "$TOOLS_JSON" > "$TEMP_JSON" && mv "$TEMP_JSON" "$TOOLS_JSON"
        
        echo "  → $FRESHNESS ($DAYS days ago)"
    fi
    
    # Rate limiting - GitHub API allows 60 requests/hour without auth
    sleep 1
done

# Update metadata
jq '.metadata.last_freshness_check = (now | todate)' "$TOOLS_JSON" > "$TEMP_JSON" && mv "$TEMP_JSON" "$TOOLS_JSON"

echo "✅ Freshness check complete!"
