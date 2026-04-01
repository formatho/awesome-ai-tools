#!/usr/bin/env node
/**
 * Freshness Checker for Awesome AI Tools
 * Updates last_commit_date and freshness status for all GitHub repos
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const REPO_DIR = path.join(__dirname, '..');
const TOOLS_JSON = path.join(REPO_DIR, 'data', 'tools.json');

// GitHub API rate limit: 60 requests/hour without auth
const RATE_LIMIT_DELAY = 1000; // 1 second between requests
const MAX_CONCURRENT = 5;

// Freshness categories
const FRESHNESS_THRESHOLDS = {
  FRESH: 7,      // ≤ 7 days
  RECENT: 30,    // ≤ 30 days
  AGING: 90,     // ≤ 90 days
  STALE: Infinity // > 90 days
};

const FRESHNESS_EMOJIS = {
  FRESH: '🟢',
  RECENT: '🟡',
  AGING: '🟠',
  STALE: '🔴'
};

function getFreshnessStatus(daysSinceCommit) {
  if (daysSinceCommit <= FRESHNESS_THRESHOLDS.FRESH) {
    return { status: 'Fresh', emoji: FRESHNESS_EMOJIS.FRESH };
  } else if (daysSinceCommit <= FRESHNESS_THRESHOLDS.RECENT) {
    return { status: 'Recent', emoji: FRESHNESS_EMOJIS.RECENT };
  } else if (daysSinceCommit <= FRESHNESS_THRESHOLDS.AGING) {
    return { status: 'Aging', emoji: FRESHNESS_EMOJIS.AGING };
  } else {
    return { status: 'Stale', emoji: FRESHNESS_EMOJIS.STALE };
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function fetchGitHubRepoInfo(url) {
  // Convert GitHub URL to API URL
  const match = url.match(/github\.com\/([^\/]+\/[^\/]+)/);
  if (!match) return null;
  
  const [owner, repo] = match[1].split('/');
  const apiUrl = `https://api.github.com/repos/${owner}/${repo}`;
  
  try {
    // Use GitHub token for higher rate limits (5000 requests/hour vs 60)
    const headers = {
      'Accept': 'application/vnd.github.v3+json',
      'User-Agent': 'Awesome-AI-Tools-Freshness-Checker'
    };
    
    // Try to get token from gh CLI or environment
    const token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN;
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(apiUrl, { headers });
    if (!response.ok) {
      if (response.status === 403) {
        console.error(`  ❌ Rate limit hit for ${owner}/${repo}. Consider adding GITHUB_TOKEN.`);
      } else {
        console.error(`  ❌ Failed to fetch ${owner}/${repo}: ${response.status}`);
      }
      return null;
    }
    return await response.json();
  } catch (error) {
    console.error(`  ❌ Error fetching ${owner}/${repo}:`, error.message);
    return null;
  }
}

function calculateDaysSinceCommit(pushedAt) {
  const commitDate = new Date(pushedAt);
  const now = new Date();
  const diffMs = now - commitDate;
  return Math.floor(diffMs / (1000 * 60 * 60 * 24));
}

async function checkFreshness() {
  console.log('🔍 Checking freshness of GitHub repos...\n');
  
  // Read tools.json
  const data = JSON.parse(fs.readFileSync(TOOLS_JSON, 'utf8'));
  const tools = data.tools;
  
  let updated = 0;
  let checked = 0;
  let errors = 0;
  
  for (const tool of tools) {
    if (!tool.url || !tool.url.includes('github.com')) {
      continue;
    }
    
    checked++;
    console.log(`Checking: ${tool.name}`);
    
    const repoInfo = await fetchGitHubRepoInfo(tool.url);
    
    if (repoInfo && repoInfo.pushed_at) {
      const daysSinceCommit = calculateDaysSinceCommit(repoInfo.pushed_at);
      const { status, emoji } = getFreshnessStatus(daysSinceCommit);
      
      tool.last_commit_date = repoInfo.pushed_at;
      tool.days_since_commit = daysSinceCommit;
      tool.freshness = `${emoji} ${status}`;
      tool.stars = repoInfo.stargazers_count; // Also update stars
      
      console.log(`  → ${emoji} ${status} (${daysSinceCommit} days ago, ${repoInfo.stargazers_count} ⭐)`);
      updated++;
    } else {
      errors++;
    }
    
    // Rate limiting
    await sleep(RATE_LIMIT_DELAY);
  }
  
  // Update metadata
  data.metadata.last_freshness_check = new Date().toISOString();
  data.metadata.last_updated = new Date().toISOString();
  
  // Write back to file with proper formatting
  const jsonContent = JSON.stringify(data, null, 2);
  fs.writeFileSync(TOOLS_JSON, jsonContent, 'utf8');
  
  console.log('\n✅ Freshness check complete!');
  console.log(`   Checked: ${checked} repos`);
  console.log(`   Updated: ${updated} tools`);
  console.log(`   Errors: ${errors}`);
  console.log(`   Total size: ${(jsonContent.length / 1024).toFixed(2)} KB`);
  
  return { checked, updated, errors };
}

// Run the checker
checkFreshness().catch(error => {
  console.error('❌ Fatal error:', error);
  process.exit(1);
});
