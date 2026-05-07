#!/usr/bin/env node
/**
 * Simple discovery script - fetch trending AI tools from GitHub
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const REPO_DIR = path.join(__dirname, '..');
const TOOLS_JSON = path.join(REPO_DIR, 'data', 'tools.json');

// Load existing tools
const data = JSON.parse(fs.readFileSync(TOOLS_JSON, 'utf8'));
const existingUrls = new Set(data.tools.map(t => t.url.toLowerCase()));
const existingNames = new Set(data.tools.map(t => t.name.toLowerCase()));

function normalizeUrl(url) {
  return url.replace('https://www.', 'https://').replace(/\/$/, '').toLowerCase();
}

function isDuplicate(name, url) {
  const normalizedUrl = normalizeUrl(url);
  if (existingUrls.has(normalizedUrl)) return true;
  if (existingNames.has(name.toLowerCase())) return true;
  return false;
}

function categorizeTool(name, description, language) {
  const lowerDesc = (description + ' ' + name).toLowerCase();

  if (lowerDesc.includes('llm') || lowerDesc.includes('chatbot') || lowerDesc.includes('gpt') ||
      lowerDesc.includes('claude') || lowerDesc.includes('chat') || lowerDesc.includes('conversational')) {
    return 'LLMs & Chatbots';
  }
  if (lowerDesc.includes('image') || lowerDesc.includes('diffusion') || lowerDesc.includes('stable') ||
      lowerDesc.includes('dall') || lowerDesc.includes('midjourney') || lowerDesc.includes('computer vision')) {
    return 'Image Generation';
  }
  if (lowerDesc.includes('video') || lowerDesc.includes('animation') || lowerDesc.includes('motion')) {
    return 'Video & Animation';
  }
  if (lowerDesc.includes('audio') || lowerDesc.includes('music') || lowerDesc.includes('speech') ||
      lowerDesc.includes('tts') || lowerDesc.includes('voice') || lowerDesc.includes('whisper')) {
    return 'Audio & Music';
  }
  if (lowerDesc.includes('agent') || lowerDesc.includes('automation') || lowerDesc.includes('autonomous')) {
    return 'Agents & Automation';
  }
  if (lowerDesc.includes('developer') || lowerDesc.includes('code') || lowerDesc.includes('ide')) {
    return 'Developer Tools';
  }
  if (lowerDesc.includes('model') || lowerDesc.includes('transformer') || lowerDesc.includes('pretrained')) {
    return 'Open Source Models';
  }
  if (lowerDesc.includes('research') || lowerDesc.includes('dataset') || lowerDesc.includes('machine learning')) {
    return 'Research & Data';
  }
  if (lowerDesc.includes('productivity') || lowerDesc.includes('workflow') || lowerDesc.includes('task')) {
    return 'Productivity';
  }

  return 'LLMs & Chatbots';
}

function extractTags(name, description, language) {
  const text = `${name} ${description}`.toLowerCase();
  const tags = [];

  const tagKeywords = [
    'ai', 'ml', 'llm', 'gpt', 'chatbot', 'automation', 'agent',
    'image', 'video', 'audio', 'text', 'code', 'api', 'open-source',
    'python', 'javascript', 'typescript', 'rust', 'go',
    'machine-learning', 'deep-learning', 'neural-network',
    'diffusion', 'transformer', 'hugging-face', 'nlp', 'computer-vision'
  ];

  tagKeywords.forEach(keyword => {
    if (text.includes(keyword) && !tags.includes(keyword)) {
      tags.push(keyword);
    }
  });

  if (language && language !== 'Unknown' && !tags.includes(language.toLowerCase())) {
    tags.push(language.toLowerCase());
  }

  return tags.slice(0, 5);
}

async function searchGitHub(query) {
  try {
    const token = execSync('gh auth token', { encoding: 'utf8' }).trim();
    const url = `https://api.github.com/search/repositories?q=${encodeURIComponent(query)}&per_page=10&sort=stars&order=desc`;

    const response = execSync(`curl -s "${url}" -H "Authorization: token ${token}"`, { encoding: 'utf8' });
    const data = JSON.parse(response);

    if (!data.items) {
      return [];
    }

    return data.items.map(item => ({
      name: item.name,
      url: item.html_url || `https://github.com/${item.full_name}`,
      description: item.description || 'No description',
      stars: item.stargazers_count,
      language: item.language || 'Unknown',
      updated_at: item.updated_at
    }));
  } catch (error) {
    console.error(`Error searching for "${query}":`, error.message);
    return [];
  }
}

async function main() {
  console.log('🔍 Discovering new AI tools from GitHub...\n');

  const queries = [
    'llm stars:>1000 pushed:>2026-04-20',
    'ai agent stars:>500 pushed:>2026-04-20',
    'machine learning stars:>1000 pushed:>2026-04-20',
    'automation stars:>500 pushed:>2026-04-20',
    'chatbot stars:>500 pushed:>2026-04-20',
    'computer vision stars:>500 pushed:>2026-04-20',
    'nlp stars:>500 pushed:>2026-04-20',
    'diffusion stars:>500 pushed:>2026-04-20',
    'transformer stars:>500 pushed:>2026-04-20'
  ];

  const newTools = [];

  for (const query of queries) {
    console.log(`Searching: ${query}`);
    const results = await searchGitHub(query);

    for (const tool of results) {
      if (!isDuplicate(tool.name, tool.url)) {
        const category = categorizeTool(tool.name, tool.description, tool.language);
        const tags = extractTags(tool.name, tool.description, tool.language);
        const daysSince = Math.floor((new Date() - new Date(tool.updated_at)) / (1000 * 60 * 60 * 24));

        newTools.push({
          name: tool.name,
          url: tool.url,
          description: tool.description,
          category,
          tags,
          pricing: 'Free',
          stars: tool.stars,
          language: tool.language,
          added_date: new Date().toISOString().split('T')[0],
          verified: true,
          freshness: '🟢 Fresh',
          last_commit_date: tool.updated_at,
          days_since_commit: daysSince
        });

        console.log(`  ✨ New: ${tool.name} (${tool.stars} ⭐) - ${category}`);
      }
    }
  }

  if (newTools.length === 0) {
    console.log('\n✅ No new tools discovered. Repository is up to date!');
    return;
  }

  console.log(`\n🎉 Discovered ${newTools.length} new tools!`);

  // Add new tools to data
  data.tools.push(...newTools);

  // Update metadata
  data.metadata.total_tools = data.tools.length;
  data.metadata.last_updated = new Date().toISOString();
  data.metadata.last_discovery = new Date().toISOString();

  // Write back
  fs.writeFileSync(TOOLS_JSON, JSON.stringify(data, null, 2), 'utf8');

  console.log(`\n✅ Added ${newTools.length} tools to database`);
  console.log(`   Total tools: ${data.metadata.total_tools}`);

  // Commit changes
  try {
    console.log('\n📝 Committing changes...');
    execSync('git add data/tools.json', { cwd: REPO_DIR, stdio: 'inherit' });
    execSync(`git commit -m "🤖 Auto-discovery: Add ${newTools.length} new trending AI tools"`, { cwd: REPO_DIR, stdio: 'inherit' });
    console.log('✅ Committed successfully');

    console.log('\n🚀 Pushing to GitHub...');
    execSync('git push origin main', { cwd: REPO_DIR, stdio: 'inherit' });
    console.log('✅ Pushed successfully');
  } catch (error) {
    console.error('❌ Error during git operations:', error.message);
  }

  // Summary
  console.log('\n📊 Discovery Summary:');
  console.log('========================================');
  newTools.forEach((tool, index) => {
    console.log(`${index + 1}. **${tool.name}** (${tool.stars} ⭐)`);
    console.log(`   Category: ${tool.category}`);
    console.log(`   URL: ${tool.url}`);
    console.log(`   ${tool.description.substring(0, 80)}...`);
    console.log('');
  });
}

main().catch(console.error);
