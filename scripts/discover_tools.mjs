#!/usr/bin/env node
/**
 * Discover new AI tools from GitHub search API
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { execSync } from 'child_process';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const REPO_DIR = path.join(__dirname, '..');
const TOOLS_JSON = path.join(REPO_DIR, 'data', 'tools.json');
const CHECK_DUPLICATES = path.join(__dirname, 'check_duplicates.mjs');

// Import isDuplicate function
const { isDuplicate } = await import(CHECK_DUPLICATES);

function getLanguage(category) {
  const languageMap = {
    'LLMs & Chatbots': 'Python',
    'Image Generation': 'Python',
    'Video & Animation': 'Python',
    'Audio & Music': 'Python',
    'Developer Tools': 'TypeScript',
    'Productivity': 'JavaScript',
    'Research & Data': 'Python',
    'Agents & Automation': 'Python',
    'Open Source Models': 'Python'
  };
  return languageMap[category] || 'Unknown';
}

function categorizeTool(name, description, language) {
  const lowerDesc = (description + ' ' + name).toLowerCase();

  const categoryRules = [
    { category: 'LLMs & Chatbots', keywords: ['llm', 'chatbot', 'gpt', 'claude', 'chat', 'conversational', 'dialogue', 'assistant'] },
    { category: 'Image Generation', keywords: ['image generation', 'diffusion', 'stable diffusion', 'dall-e', 'midjourney', 'image ai', 'computer vision'] },
    { category: 'Video & Animation', keywords: ['video', 'animation', 'motion', 'ffmpeg', 'video ai'] },
    { category: 'Audio & Music', keywords: ['audio', 'music', 'speech', 'tts', 'voice', 'sound', 'whisper'] },
    { category: 'Developer Tools', keywords: ['developer', 'code', 'ide', 'editor', 'debug', 'testing', 'api'] },
    { category: 'Productivity', keywords: ['productivity', 'automation', 'workflow', 'task', 'note', 'calendar'] },
    { category: 'Research & Data', keywords: ['research', 'data', 'dataset', 'machine learning', 'deep learning', 'neural network', 'model'] },
    { category: 'Agents & Automation', keywords: ['agent', 'automation', 'autonomous', 'robot', 'workflow', 'pipeline'] },
    { category: 'Open Source Models', keywords: ['model', 'pretrained', 'weights', 'checkpoint', 'hugging face', 'transformer'] }
  ];

  for (const rule of categoryRules) {
    for (const keyword of rule.keywords) {
      if (lowerDesc.includes(keyword)) {
        return rule.category;
      }
    }
  }

  // Default category based on language
  if (['Python', 'R', 'Julia'].includes(language)) {
    return 'Research & Data';
  } else if (['JavaScript', 'TypeScript'].includes(language)) {
    return 'Developer Tools';
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
    if (text.includes(keyword)) {
      tags.push(keyword);
    }
  });

  // Add language as tag
  if (language && language !== 'Unknown') {
    tags.push(language.toLowerCase());
  }

  // Limit to top 5 tags
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

async function discoverNewTools() {
  console.log('🔍 Discovering new AI tools from GitHub...\n');

  const queries = [
    'llm stars:>1000 pushed:>2026-04-20',
    'ai stars:>1000 pushed:>2026-04-20',
    'machine learning stars:>1000 pushed:>2026-04-20',
    'automation stars:>500 pushed:>2026-04-20',
    'chatbot stars:>500 pushed:>2026-04-20',
    'computer vision stars:>500 pushed:>2026-04-20',
    'nlp stars:>500 pushed:>2026-04-20',
    'diffusion stars:>500 pushed:>2026-04-20',
    'agent stars:>500 pushed:>2026-04-20',
    'transformer stars:>500 pushed:>2026-04-20'
  ];

  const discoveredTools = [];

  for (const query of queries) {
    console.log(`Searching: ${query}`);
    const results = await searchGitHub(query);

    for (const tool of results) {
      const duplicateCheck = isDuplicate(tool.name, tool.url);

      if (!duplicateCheck.duplicate) {
        const category = categorizeTool(tool.name, tool.description, tool.language);
        const tags = extractTags(tool.name, tool.description, tool.language);

        discoveredTools.push({
          name: tool.name,
          url: tool.url,
          description: tool.description,
          category,
          tags: JSON.stringify(tags),
          pricing: 'Free',
          stars: tool.stars,
          language: tool.language,
          added_date: new Date().toISOString().split('T')[0],
          verified: true,
          freshness: '🟢 Fresh',
          last_commit_date: tool.updated_at,
          days_since_commit: Math.floor((new Date() - new Date(tool.updated_at)) / (1000 * 60 * 60 * 24))
        });

        console.log(`  ✨ New: ${tool.name} (${tool.stars} ⭐) - ${category}`);
      }
    }
  }

  return discoveredTools;
}

async function main() {
  const newTools = await discoverNewTools();

  if (newTools.length === 0) {
    console.log('\n✅ No new tools discovered. Repository is up to date!');
    return;
  }

  console.log(`\n🎉 Discovered ${newTools.length} new tools!`);

  // Load existing tools
  const data = JSON.parse(fs.readFileSync(TOOLS_JSON, 'utf8'));

  // Add new tools
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
    execSync('git add data/tools.json', { cwd: REPO_DIR });
    execSync(`git commit -m "🤖 Auto-discovery: Add ${newTools.length} new trending AI tools"`, { cwd: REPO_DIR });
    console.log('✅ Committed successfully');

    console.log('\n🚀 Pushing to GitHub...');
    execSync('git push origin main', { cwd: REPO_DIR });
    console.log('✅ Pushed successfully');
  } catch (error) {
    console.error('❌ Error during git operations:', error.message);
  }

  // Summary
  console.log('\n📊 Discovery Summary:');
  console.log('----------------------------------------');
  newTools.forEach((tool, index) => {
    console.log(`${index + 1}. **${tool.name}** (${tool.stars} ⭐)`);
    console.log(`   Category: ${tool.category}`);
    console.log(`   URL: ${tool.url}`);
    console.log(`   Description: ${tool.description.substring(0, 80)}...`);
    console.log('');
  });
}

main().catch(console.error);
