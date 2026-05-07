#!/usr/bin/env node

/**
 * Discover trending AI/ML tools from GitHub and add them to the tools database
 */

import { readFileSync, writeFileSync } from 'fs';
import { execSync } from 'child_process';

const TOOLS_FILE = 'data/tools.json';
const GITHUB_TOKEN = process.env.GITHUB_TOKEN || execSync('gh auth token').toString().trim();

// Keywords to search for AI/ML trending repos
const SEARCH_QUERIES = [
  'topic:ai+language:python',
  'topic:llm+language:python',
  'topic:agent+language:python',
  'topic:automation+language:typescript',
  'topic:machine-learning+language:python',
  'topic:deep-learning+language:python',
  'topic:chatbot+language:python',
  'topic:image-generation+language:python',
  'topic:copilot+language:typescript',
  'topic:rag+language:python',
];

async function fetchFromGitHub(endpoint) {
  const response = await fetch(`https://api.github.com${endpoint}`, {
    headers: {
      'Authorization': `Bearer ${GITHUB_TOKEN}`,
      'Accept': 'application/vnd.github.v3+json',
      'User-Agent': 'awesome-ai-discovery'
    }
  });
  
  if (!response.ok) {
    const error = await response.text();
    throw new Error(`GitHub API error: ${response.status} - ${error}`);
  }
  
  return response.json();
}

async function searchRepositories(query, stars = 5000) {
  try {
    const data = await fetchFromGitHub(
      `/search/repositories?q=${encodeURIComponent(query)}+stars:>${stars}&sort=stars&order=desc&per_page=10`
    );
    return data.items || [];
  } catch (error) {
    console.error(`Error searching for "${query}":`, error.message);
    return [];
  }
}

async function getRepositoryDetails(owner, repo) {
  try {
    const [repoData, commitsData] = await Promise.all([
      fetchFromGitHub(`/repos/${owner}/${repo}`),
      fetchFromGitHub(`/repos/${owner}/${repo}/commits?per_page=1`)
    ]);
    
    return {
      ...repoData,
      last_commit: commitsData[0]?.commit?.committer?.date || null
    };
  } catch (error) {
    console.error(`Error fetching details for ${owner}/${repo}:`, error.message);
    return null;
  }
}

function loadTools() {
  try {
    const content = readFileSync(TOOLS_FILE, 'utf8');
    const data = JSON.parse(content);
    return data.tools || [];
  } catch (error) {
    console.error('Error loading tools:', error.message);
    return [];
  }
}

function saveTools(tools, metadata) {
  const data = {
    metadata,
    tools
  };
  writeFileSync(TOOLS_FILE, JSON.stringify(data, null, 2));
}

function isDuplicate(url, existingTools) {
  const normalizedUrl = url.toLowerCase().replace(/https?:\/\//, '').replace(/\/$/, '');
  
  for (const tool of existingTools) {
    const existingUrl = tool.url.toLowerCase().replace(/https?:\/\//, '').replace(/\/$/, '');
    if (normalizedUrl === existingUrl) {
      return true;
    }
    
    // Check by repo name (owner/repo)
    const repoMatch = url.match(/github\.com\/([^\/]+)\/([^\/]+)/);
    if (repoMatch) {
      const [, owner, repo] = repoMatch;
      const existingRepoMatch = tool.url.match(/github\.com\/([^\/]+)\/([^\/]+)/);
      if (existingRepoMatch) {
        const [, existingOwner, existingRepo] = existingRepoMatch;
        if (owner.toLowerCase() === existingOwner.toLowerCase() && 
            repo.toLowerCase() === existingRepo.toLowerCase()) {
          return true;
        }
      }
    }
  }
  
  return false;
}

function categorizeTool(repo) {
  const topics = repo.topics || [];
  const description = repo.description?.toLowerCase() || '';
  const name = repo.name?.toLowerCase() || '';
  
  // Category mapping based on keywords
  const categoryMap = {
    'Agents & Automation': ['agent', 'automation', 'workflow', 'copilot', 'assistant', 'bot', 'claude', 'codex', 'cursor'],
    'LLMs & Chatbots': ['llm', 'chatbot', 'chat', 'gpt', 'language model', 'nlp'],
    'Image Generation': ['image', 'stable diffusion', 'diffusion', 'generation', 'visual', 'comfyui', 'midjourney'],
    'Video & Animation': ['video', 'animation', 'deepfake', 'live cam', 'avatar'],
    'Audio & Music': ['audio', 'music', 'speech', 'voice', 'tts', 'stt'],
    'Open Source Models': ['model', 'transformer', 'bert', 'llama', 'mistral', 'pytorch', 'tensorflow'],
    'Developer Tools': ['sdk', 'api', 'library', 'framework', 'tools', 'developer', 'code'],
    'Productivity': ['productivity', 'prompt', 'assistant', 'notes', 'organize'],
    'Research & Data': ['research', 'data', 'science', 'ml', 'machine learning', 'deep learning', 'tutorial', 'course']
  };
  
  for (const [category, keywords] of Object.entries(categoryMap)) {
    for (const keyword of keywords) {
      if (name.includes(keyword) || description.includes(keyword) || topics.some(t => t.includes(keyword))) {
        return category;
      }
    }
  }
  
  return 'Developer Tools'; // Default category
}

function calculateFreshness(lastCommitDate) {
  if (!lastCommitDate) return '🟡 Recent';
  
  const now = new Date();
  const commitDate = new Date(lastCommitDate);
  const daysSinceCommit = Math.floor((now - commitDate) / (1000 * 60 * 60 * 24));
  
  if (daysSinceCommit <= 7) return '🟢 Fresh';
  if (daysSinceCommit <= 30) return '🟡 Recent';
  if (daysSinceCommit <= 90) return '🟠 Aging';
  return '🔴 Stale';
}

async function discoverAndAddTools() {
  console.log('🔍 Starting AI tool discovery...');
  
  const existingTools = loadTools();
  console.log(`📚 Loaded ${existingTools.length} existing tools`);
  
  const newTools = [];
  const seenUrls = new Set();
  
  for (const query of SEARCH_QUERIES) {
    console.log(`\n🔎 Searching: ${query}`);
    
    const repos = await searchRepositories(query);
    
    for (const repo of repos) {
      if (!repo.html_url) continue;
      
      // Check duplicate
      if (isDuplicate(repo.html_url, existingTools) || seenUrls.has(repo.html_url)) {
        continue;
      }
      
      seenUrls.add(repo.html_url);
      
      // Get detailed info
      const details = await getRepositoryDetails(repo.owner.login, repo.name);
      if (!details) continue;
      
      // Categorize
      const category = categorizeTool(details);
      const freshness = calculateFreshness(details.last_commit);
      
      // Create tool entry
      const tool = {
        name: details.name,
        url: details.html_url,
        description: details.description || '',
        category: category,
        tags: (details.topics || []).slice(0, 5),
        pricing: 'Free',
        stars: details.stargazers_count || 0,
        language: details.language || 'Unknown',
        added_date: new Date().toISOString().split('T')[0],
        verified: true,
        freshness: freshness,
        last_commit_date: details.last_commit,
        days_since_commit: details.last_commit ? 
          Math.floor((new Date() - new Date(details.last_commit)) / (1000 * 60 * 60 * 24)) : null
      };
      
      newTools.push(tool);
      console.log(`  ✅ Found: ${tool.name} (${tool.stars} stars) - ${category}`);
    }
    
    // Rate limit delay
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  if (newTools.length === 0) {
    console.log('\n✅ No new tools found');
    return;
  }
  
  console.log(`\n🎉 Found ${newTools.length} new tools!`);
  
  // Update tools file
  const updatedTools = [...newTools, ...existingTools];
  const metadata = {
    version: '1.0.0',
    last_updated: new Date().toISOString(),
    total_tools: updatedTools.length,
    categories: [
      'LLMs & Chatbots',
      'Image Generation',
      'Video & Animation',
      'Audio & Music',
      'Developer Tools',
      'Productivity',
      'Research & Data',
      'Agents & Automation',
      'Open Source Models'
    ],
    last_freshness_check: new Date().toISOString(),
    last_discovery: new Date().toISOString()
  };
  
  saveTools(updatedTools, metadata);
  console.log(`\n💾 Saved ${updatedTools.length} total tools to ${TOOLS_FILE}`);
  
  // Return summary
  return newTools.map(t => ({
    name: t.name,
    category: t.category,
    stars: t.stars,
    description: t.description
  }));
}

// Run discovery
discoverAndAddTools()
  .then(summary => {
    if (summary && summary.length > 0) {
      console.log('\n📊 Discovery Summary:');
      console.log('==================');
      summary.forEach(tool => {
        console.log(`\n${tool.name}`);
        console.log(`  Category: ${tool.category}`);
        console.log(`  Stars: ${tool.stars.toLocaleString()}`);
        console.log(`  Description: ${tool.description.substring(0, 100)}...`);
      });
      console.log(`\nTotal new tools: ${summary.length}`);
    }
    process.exit(0);
  })
  .catch(error => {
    console.error('❌ Discovery failed:', error);
    process.exit(1);
  });
