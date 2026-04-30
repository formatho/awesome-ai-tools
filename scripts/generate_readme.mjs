#!/usr/bin/env node
/**
 * README Generator for Awesome AI Tools
 * Generates README.md from tools.json with freshness indicators
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const REPO_DIR = path.join(__dirname, '..');
const TOOLS_JSON = path.join(REPO_DIR, 'data', 'tools.json');
const README_MD = path.join(REPO_DIR, 'README.md');

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
}

function generateCategorySection(tools, categoryName) {
  if (tools.length === 0) return '';
  
  let section = `## ${categoryName}\n\n`;
  section += `> **Freshness Legend:** 🟢 Fresh (≤7d) | 🟡 Recent (≤30d) | 🟠 Aging (≤90d) | 🔴 Stale (>90d)\n\n`;
  section += `| Tool | Description | Tags | Pricing | Freshness | Stars |\n`;
  section += `|------|-------------|------|---------|-----------|-------|\n`;
  
  tools.forEach(tool => {
    // Handle missing or malformed tags
    const tags = Array.isArray(tool.tags) ? tool.tags.slice(0, 3).map(t => `\`${t}\``).join(' ') : 'N/A';
    const freshness = tool.freshness || '❓ Unknown';
    const stars = tool.stars ? `${(tool.stars / 1000).toFixed(1)}k` : 'N/A';
    const pricing = tool.pricing || 'Unknown';
    
    section += `| [${tool.name}](${tool.url}) | ${tool.description} | ${tags} | ${pricing} | ${freshness} | ${stars} |\n`;
  });
  
  section += `\n---\n\n`;
  return section;
}

function generateReadme(data) {
  const { metadata, tools } = data;
  
  // Group tools by category
  const categories = {};
  tools.forEach(tool => {
    const category = tool.category || 'Other';
    if (!categories[category]) {
      categories[category] = [];
    }
    categories[category].push(tool);
  });
  
  // Sort tools within each category by stars (descending)
  Object.keys(categories).forEach(category => {
    categories[category].sort((a, b) => (b.stars || 0) - (a.stars || 0));
  });
  
  // Generate README content
  let readme = `# 🤖 Awesome AI Tools\n\n`;
  readme += `> A continuously updated, curated list of high-quality AI tools and resources.\n\n`;
  
  readme += `[![Last Updated](https://img.shields.io/badge/Updated-${new Date().toISOString().split('T')[0]}-blue)](https://github.com/formatho/awesome-ai-tools)\n`;
  readme += `[![Total Tools](https://img.shields.io/badge/Tools-${metadata.total_tools}-green)](data/tools.json)\n`;
  readme += `[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)\n\n`;
  
  readme += `## 📋 Table of Contents\n\n`;
  readme += `- [Recently Added](#-recently-added)\n`;
  Object.keys(categories).sort().forEach(category => {
    readme += `- [${category}](#${category.toLowerCase().replace(/ /g, '-').replace(/&/g, '-')})\n`;
  });
  readme += `- [Trending Tools](#-trending-tools)\n`;
  readme += `- [Statistics](#-statistics)\n`;
  readme += `- [Contributing](#-contributing)\n\n`;
  readme += `---\n\n`;
  
  // Recently Added (last 10 tools)
  readme += `## 🆕 Recently Added\n\n`;
  readme += `| Tool | Description | Category | Freshness | Added |\n`;
  readme += `|------|-------------|----------|-----------|-------|\n`;
  
  const recentTools = tools
    .filter(t => t.added_date)
    .sort((a, b) => new Date(b.added_date) - new Date(a.added_date))
    .slice(0, 10);
  
  recentTools.forEach(tool => {
    const freshness = tool.freshness || '❓';
    const addedDate = formatDate(tool.added_date);
    readme += `| [${tool.name}](${tool.url}) | ${tool.description} | ${tool.category} | ${freshness} | ${addedDate} |\n`;
  });
  
  readme += `\n---\n\n`;
  
  // Category sections
  Object.keys(categories).sort().forEach(categoryName => {
    readme += generateCategorySection(categories[categoryName], categoryName);
  });
  
  // Trending Tools (top 10 by stars)
  readme += `## 📈 Trending Tools\n\n`;
  readme += `Top 10 by GitHub stars:\n\n`;
  
  const trending = tools
    .filter(t => t.stars)
    .sort((a, b) => b.stars - a.stars)
    .slice(0, 10);
  
  trending.forEach((tool, index) => {
    const stars = (tool.stars / 1000).toFixed(1);
    const freshness = tool.freshness || '❓';
    readme += `${index + 1}. **[${tool.name}](${tool.url})** - ${stars}k ⭐ ${freshness} - ${tool.description}\n`;
  });
  
  readme += `\n---\n\n`;
  
  // Statistics
  readme += `## 📊 Statistics\n\n`;
  readme += `- **Total Tools:** ${metadata.total_tools}\n`;
  readme += `- **Categories:** ${metadata.categories.length}\n`;
  readme += `- **Last Updated:** ${formatDate(metadata.last_updated)}\n`;
  readme += `- **Last Freshness Check:** ${metadata.last_freshness_check ? formatDate(metadata.last_freshness_check) : 'Never'}\n`;
  readme += `- **Update Frequency:** Every 6 hours (discovery) + Daily (freshness)\n\n`;
  
  // Freshness distribution
  const freshCount = tools.filter(t => t.freshness && t.freshness.includes('Fresh')).length;
  const recentCount = tools.filter(t => t.freshness && t.freshness.includes('Recent')).length;
  const agingCount = tools.filter(t => t.freshness && t.freshness.includes('Aging')).length;
  const staleCount = tools.filter(t => t.freshness && t.freshness.includes('Stale')).length;
  
  readme += `**Freshness Distribution:**\n`;
  readme += `- 🟢 Fresh: ${freshCount} tools\n`;
  readme += `- 🟡 Recent: ${recentCount} tools\n`;
  readme += `- 🟠 Aging: ${agingCount} tools\n`;
  readme += `- 🔴 Stale: ${staleCount} tools\n\n`;
  
  readme += `---\n\n`;
  
  // License & Footer
  readme += `## 📄 License\n\n`;
  readme += `This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.\n\n`;
  readme += `---\n\n`;
  readme += `**Maintained by [Formatho](https://formatho.com)** | Autonomous AI Research Agent 🤖\n`;
  
  return readme;
}

// Generate and write README
const data = JSON.parse(fs.readFileSync(TOOLS_JSON, 'utf8'));
const readme = generateReadme(data);
fs.writeFileSync(README_MD, readme, 'utf8');

console.log('✅ README.md generated successfully!');
console.log(`   Total tools: ${data.metadata.total_tools}`);
console.log(`   Categories: ${data.metadata.categories.length}`);
