#!/usr/bin/env node
/**
 * Test Freshness Checker - Check first 5 tools only
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const REPO_DIR = path.join(__dirname, '..');
const TOOLS_JSON = path.join(REPO_DIR, 'data', 'tools.json');

const FRESHNESS_THRESHOLDS = {
  FRESH: 7,
  RECENT: 30,
  AGING: 90,
  STALE: Infinity
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
  const match = url.match(/github\.com\/([^\/]+\/[^\/]+)/);
  if (!match) return null;
  
  const [owner, repo] = match[1].split('/');
  const apiUrl = `https://api.github.com/repos/${owner}/${repo}`;
  
  try {
    const response = await fetch(apiUrl);
    if (!response.ok) {
      console.error(`  ❌ Failed: ${response.status}`);
      return null;
    }
    return await response.json();
  } catch (error) {
    console.error(`  ❌ Error:`, error.message);
    return null;
  }
}

function calculateDaysSinceCommit(pushedAt) {
  const commitDate = new Date(pushedAt);
  const now = new Date();
  const diffMs = now - commitDate;
  return Math.floor(diffMs / (1000 * 60 * 60 * 24));
}

async function testFreshness() {
  console.log('🧪 Testing freshness checker (first 5 GitHub repos)...\n');
  
  const data = JSON.parse(fs.readFileSync(TOOLS_JSON, 'utf8'));
  const githubTools = data.tools.filter(t => t.url && t.url.includes('github.com')).slice(0, 5);
  
  for (const tool of githubTools) {
    console.log(`Checking: ${tool.name}`);
    
    const repoInfo = await fetchGitHubRepoInfo(tool.url);
    
    if (repoInfo && repoInfo.pushed_at) {
      const daysSinceCommit = calculateDaysSinceCommit(repoInfo.pushed_at);
      const { status, emoji } = getFreshnessStatus(daysSinceCommit);
      
      console.log(`  ${emoji} ${status} (${daysSinceCommit} days ago)`);
      console.log(`  ⭐ Stars: ${repoInfo.stargazers_count}`);
      console.log(`  📅 Last commit: ${repoInfo.pushed_at}\n`);
    }
    
    await sleep(1000);
  }
  
  console.log('✅ Test complete!');
}

testFreshness().catch(error => {
  console.error('❌ Error:', error);
  process.exit(1);
});
