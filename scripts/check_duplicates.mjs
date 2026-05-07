#!/usr/bin/env node
/**
 * Check if a repository URL already exists in tools.json
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const REPO_DIR = path.join(__dirname, '..');
const TOOLS_JSON = path.join(REPO_DIR, 'data', 'tools.json');

function normalizeUrl(url) {
  if (!url || typeof url !== 'string') {
    return '';
  }
  // Normalize URL by removing 'www', 'api.github.com/repos', and trailing slashes
  return url
    .replace('https://www.', 'https://')
    .replace('https://api.github.com/repos/', 'https://github.com/')
    .replace('https://github.com/github.com/', 'https://github.com/')
    .replace(/\/$/, '')
    .toLowerCase();
}

function extractRepoOwner(url) {
  const match = url.match(/github\.com\/([^\/]+)\/([^\/]+)/);
  if (match) {
    return `${match[1]}/${match[2]}`;
  }
  return null;
}

// Load existing tools
const data = JSON.parse(fs.readFileSync(TOOLS_JSON, 'utf8'));
const existingUrls = new Set();
const existingRepos = new Set();
const existingNames = new Set();

data.tools.forEach(tool => {
  // Skip if tool is not an object or is an array
  if (!tool || typeof tool !== 'object' || Array.isArray(tool)) {
    return;
  }
  
  if (tool.url) {
    existingUrls.add(normalizeUrl(tool.url));
    const repo = extractRepoOwner(tool.url);
    if (repo) {
      existingRepos.add(repo.toLowerCase());
    }
  }
  
  if (tool.name) {
    existingNames.add(tool.name.toLowerCase());
  }
});

// Function to check if a tool is duplicate
function isDuplicate(name, url) {
  const normalizedUrl = normalizeUrl(url);
  const repo = extractRepoOwner(url);
  const normalizedName = name.toLowerCase();

  if (existingUrls.has(normalizedUrl)) {
    return { duplicate: true, reason: 'URL match' };
  }

  if (repo && existingRepos.has(repo.toLowerCase())) {
    return { duplicate: true, reason: 'Repository match' };
  }

  if (existingNames.has(normalizedName)) {
    return { duplicate: true, reason: 'Name match' };
  }

  return { duplicate: false, reason: 'New' };
}

// Export for use as module
export { isDuplicate };

// CLI usage
if (import.meta.url === `file://${process.argv[1]}`) {
  const url = process.argv[2];
  const name = process.argv[3] || path.basename(url);

  if (!url) {
    console.log('Usage: node check_duplicates.mjs <url> [name]');
    process.exit(1);
  }

  const result = isDuplicate(name, url);
  console.log(JSON.stringify(result, null, 2));
}
