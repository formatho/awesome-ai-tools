#!/bin/bash
# Discover new AI tools using GitHub CLI

cd "$(dirname "$0")/.."

echo "🔍 Searching for new AI/ML repositories..."

# Search queries for different categories
declare -a queries=(
  "language:python topic:ai stars:>500 pushed:>2026-04-23"
  "language:python topic:ml stars:>500 pushed:>2026-04-23"
  "language:python topic:llm stars:>500 pushed:>2026-04-23"
  "language:javascript topic:ai stars:>500 pushed:>2026-04-23"
  "language:typescript topic:ai stars:>500 pushed:>2026-04-23"
  "language:rust topic:ai stars:>500 pushed:>2026-04-23"
  "language:go topic:ai stars:>500 pushed:>2026-04-23"
  "topic:automation stars:>500 pushed:>2026-04-23"
  "topic:chatbot stars:>500 pushed:>2026-04-23"
  "topic:computer-vision stars:>500 pushed:>2026-04-23"
)

# Array to store discovered repos
declare -a discovered_repos

for query in "${queries[@]}"; do
  echo "Searching: $query"
  results=$(gh search repos --limit 20 --json name,url,description,stargazersCount,languages,updatedAt -- "$query" 2>/dev/null || echo "[]")

  # Extract repo URLs from results
  if [[ "$results" != "[]" ]]; then
    echo "$results" | jq -r '.[] | .url' >> /tmp/discovered_repos.txt
  fi
done

# Remove duplicates and sort
if [[ -f /tmp/discovered_repos.txt ]]; then
  sort -u /tmp/discovered_repos.txt > /tmp/unique_repos.txt
  echo "📊 Found $(wc -l < /tmp/unique_repos.txt) unique repositories"
  cat /tmp/unique_repos.txt
else
  echo "⚠️ No repositories found"
fi
