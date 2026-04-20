# Research Log - Formatho Blockchain/Infrastructure Partnership Research

## 2026-04-19 - Initialization & Bootstrap

**18:51 UTC** - Project initialized via cron
- Created workspace structure
- Set up research-state.yaml with bootstrap status
- Research question: What blockchain and infrastructure partnerships should Formatho pursue for growth?

**19:00 UTC** - Bootstrap Complete
- Conducted literature survey on 3 major blockchain infrastructure providers
- Analyzed Infura, QuickNode, and Alchemy partnership models
- Identified 6 key partnership models: Marketplace, Referral, Startup Program, Case Study, Technical Integration, Enterprise Co-selling
- Formed 4 initial hypotheses prioritized by time-to-results
- Transitioned from bootstrap to inner loop

**Literature sources:**
- Infura (Consenys): Technical integration focus, MetaMask ecosystem
- QuickNode: Structured partnership program with marketplace, referrals, startup program
- Alchemy: Enterprise focus, startup program with $20M Solana fund

**Hypotheses prioritized:**
1. H1: QuickNode Marketplace Integration (7 days, ≥5 signups)
2. H3: QuickNode Referral Partnership (30 days, ≥$100 revenue)
3. H2: Alchemy Startup Program (14 days, acceptance + credits)
4. H4: Case Study Partnership (14 days, ≥3 inbound leads)

**Decision rationale:** QuickNode prioritized due to structured marketplace, clear revenue path, faster time-to-results. Alchemy for enterprise follow-up.

**Next steps:**
1. Update research state with protocol completion
2. Commit progress to git
3. Conduct additional research on H2 (Alchemy) and H3 (Referral) in parallel
4. Prepare execution plan for when Formatho access is available

**19:15 UTC** - H1 Research Complete, Execution Blocked
- Analyzed QuickNode marketplace structure in detail
- Documented successful examples (GoldRush by Covalent, 0x Swap API)
- Identified optimal category: Data & Analytics + 🤖 AI Enabled badge
- Application URL found: https://quiknode.typeform.com/to/iUWe13BB
- Pricing strategy defined: FREE + $49-99/month
- **Constraint identified:** Cannot submit application without Formatho product access/credentials
- **Decision:** Pivot to researching H2 and H3 in parallel while H1 waits for access

**Key findings from marketplace analysis:**
- External providers (Covalent, 0x) succeed without being QuickNode-built
- AI category exists and is active
- Freemium pricing drives adoption
- Staff Pick/Top Seller badges boost visibility significantly

## 2026-04-20 - Deep Dive & Critical Findings

**04:50 UTC** - H3 CRITICAL FINDING: Referral Program Invalidated
- **Major issue discovered:** QuickNode referral program generates ACCOUNT CREDIT, not revenue
- **Problem:** Credits only usable if Formatho is already a QuickNode customer
- **Implication:** If Formatho doesn't use QuickNode, this hypothesis is worthless
- **Decision:** Invalidate H3, lower priority to 4, deprioritize
- **Updated hypothesis priorities:**
  1. H4: Case Study (can execute without access)
  2. H1: QuickNode Marketplace (best for revenue, but blocked)
  2. H2: Alchemy Startup (high value, can execute when access available)
  3. H4: Case Study Partnership (highest priority for no-access execution)
  4. H3: Referral Partnership (invalidated - not revenue)

**05:00 UTC** - H2 Research Complete
- Analyzed Alchemy Startup Program in detail
- **Application:** 30-second form, minimal information required
- **Benefits:** Alchemy credits + $5,000 AWS credits + 24/7 support + partner ecosystem
- **Eligibility:** "Any team building onchain" - very broad
- **Success probability:** High (70-80%)
- **Timeline:** "Credits will hit your account within days" - realistic 14-day goal
- **Value:** Conservative estimate $6,000-10,000 in total value
- **Decision:** Keep H2 as priority 2, ready to execute when Formatho access available

**05:15 UTC** - H4 Protocol Created
- **Created:** experiments/H4-case-study/protocol.md
- **Reasoning:** Case study partnerships don't require technical integration
- **Key advantages:**
  - Can execute without Formatho product access
  - Scalable once process validated
  - Builds social proof and credibility
  - Cross-pollinates audiences with partners
- **Methodology:** 4-phase approach (partner identification → case study creation → publication → lead capture)
- **Target:** ≥3 inbound leads within 14 days
- **Distribution channels:** Partner blogs, Formatho blog, LinkedIn, Twitter, Reddit, Dev.to, Medium
- **Tracking:** UTM parameters, Google Analytics, lead source attribution
- **Decision:** H4 is now PRIORITY 1 - best path to revenue without product access

**05:30 UTC** - Research State Updated
- Updated research-state.yaml with new findings
- Changed H1 status to "blocked" with documented blocker
- Changed H3 status to "invalidated" with reason
- Changed H2 status to "researched"
- Created H4 protocol, status "researching"
- Updated current_direction to focus on no-access execution strategies
- Updated priorities based on access constraints

**Current Status:**
- Literature survey: ✅ Complete
- H1 (QuickNode Marketplace): ✅ Research complete, ❌ Execution blocked
- H2 (Alchemy Startup): ✅ Research complete, ⏳ Ready when access available
- H3 (QuickNode Referral): ✅ Research complete, ⚠️ Invalidated (not revenue)
- H4 (Case Study): ✅ Protocol created, 🔍 Researching execution strategy

**Strategic Pivot:** Focus on partnerships that don't require Formatho product access:
1. **Immediate (no access needed):** H4 Case Study partnerships
2. **When access available:** H1 Marketplace integration, H2 Startup program
3. **Deprioritized:** H3 Referral program (not revenue, only valuable if using QuickNode)

**Next actions:**
1. Deep dive on H4 execution strategy (research successful case studies, identify partners)
2. Research additional no-access partnership models
3. Prepare execution plan for H4
4. Generate comprehensive progress report for human
