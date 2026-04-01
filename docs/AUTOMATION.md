# Automation & Cron Jobs

This document describes the automated jobs configured for the awesome-ai-tools repository.

## 🤖 Automated Jobs

### 1. Discovery Cycle (Every 6 Hours)

**Job ID:** `38f35496-4cae-41b5-ae9c-198774214f15`  
**Name:** `awesome-ai-discovery`  
**Schedule:** Every 6 hours  
**Agent:** `awesome-ai` (isolated session)

**Workflow:**
1. Discovery Phase - Search for new AI tools using:
   - GitHub trending (AI-related repos)
   - GitHub search (AI tools, LLMs, agents)
   - Product Hunt (latest launches)
   - Reddit (r/AItools, r/MachineLearning, r/Artificial)
   - Hacker News
   - Twitter/X trending AI discussions

2. Validation Phase - Filter tools by:
   - Working links
   - Real AI functionality
   - No spam/low-effort wrappers
   - Documentation available
   - Active development

3. Enrichment Phase - Add metadata:
   - Clean description (1-2 lines)
   - Category classification
   - Tags (LLM, image-gen, open-source, API, etc.)
   - Pricing model (Free/Freemium/Paid)

4. Deduplication - Compare against existing tools:
   - Name similarity
   - URL matching

5. Classification - Organize into 9 categories:
   - LLMs & Chatbots
   - Image Generation
   - Video & Animation
   - Audio & Music
   - Developer Tools
   - Productivity
   - Research & Data
   - Agents & Automation
   - Open Source Models

6. Repository Update:
   - Update `data/tools.json`
   - Regenerate `README.md`
   - Commit with message: `feat: add new AI tools (auto-update)`
   - Push to GitHub

**Manual Trigger:**
```bash
openclaw cron run 38f35496-4cae-41b5-ae9c-198774214f15
```

---

### 2. Freshness Check (Daily)

**Job ID:** `63b3b8de-5d0b-4988-8790-923666679f90`  
**Name:** `awesome-ai-freshness-check`  
**Schedule:** Every 24 hours (approx. 02:00 IST)  
**Agent:** Main session (Premchand)

**Workflow:**
1. Fetch last commit date from GitHub API for all repos
2. Calculate days since last commit
3. Update freshness status:
   - 🟢 **Fresh** - Updated ≤ 7 days ago
   - 🟡 **Recent** - Updated ≤ 30 days ago
   - 🟠 **Aging** - Updated ≤ 90 days ago
   - 🔴 **Stale** - Not updated in 90+ days

4. Update star counts
5. Regenerate `README.md` with freshness indicators
6. Commit with message: `chore: daily freshness update`
7. Push to GitHub

**Manual Trigger:**
```bash
openclaw cron run 63b3b8de-5d0b-4988-8790-923666679f90
```

**Manual Execution:**
```bash
cd /Users/studio/sandbox/formatho/awesome-ai-tools
GITHUB_TOKEN=$(gh auth token) node scripts/check_freshness.mjs
node scripts/generate_readme.mjs
git add . && git commit -m "chore: manual freshness update"
git push origin main
```

---

## 📊 Current Status

**Total Tools:** 66  
**Categories:** 9  
**Update Frequency:** 
- Discovery: Every 6 hours
- Freshness: Every 24 hours

**Freshness Distribution:**
- 🟢 Fresh: 42 tools (62%)
- 🟡 Recent: 4 tools (6%)
- 🟠 Aging: 10 tools (15%)
- 🔴 Stale: 12 tools (18%)

---

## 🔧 Management Commands

### View All Cron Jobs
```bash
openclaw cron list
```

### View Job Details
```bash
openclaw cron list --json
```

### Run Job Manually
```bash
openclaw cron run <job-id>
```

### Edit Job
```bash
openclaw cron edit <job-id>
```

### Disable Job
```bash
openclaw cron disable <job-id>
```

### Enable Job
```bash
openclaw cron enable <job-id>
```

### Delete Job
```bash
openclaw cron rm <job-id>
```

### View Run History
```bash
openclaw cron runs <job-id>
```

---

## 📁 Related Files

- `data/tools.json` - Master database of tools
- `data/metadata.json` - Tracking & timestamps
- `scripts/check_freshness.mjs` - Freshness checker
- `scripts/generate_readme.mjs` - README generator
- `README.md` - Human-readable list (auto-generated)

---

## 🚨 Notes

- GitHub API rate limit: 5000 requests/hour with token (vs 60/hour without)
- Jobs use `GITHUB_TOKEN=$(gh auth token)` for authentication
- README is regenerated after every freshness check
- All changes are automatically committed and pushed to GitHub

---

**Last Updated:** April 2, 2026  
**Maintained by:** Premchand 🏗️ (Autonomous AI Research Agent)
