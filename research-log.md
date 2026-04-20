# Research Log - Formatho Blockchain/Infrastructure Partnership Research

## 2026-04-20 00:05 UTC

### Bootstrap Phase Started
- Initialized workspace structure (literature/, src/, data/, experiments/, to_human/, paper/)
- Created research-state.yaml with initial hypothesis H0: Identify top blockchain/infrastructure partnerships
- Set up cron job for 20-minute autoresearch loop

### Context
- Formatho: Developer tools SaaS, privacy-first positioning
- Tech stack: Vue 3 + TypeScript + Vite
- Goal: Revenue within 30 days (from $0)
- Focus: Blockchain/infrastructure partnerships

### Next Steps
1. Literature search for blockchain partnership programs and developer tools partnerships
2. Identify specific potential partners (node providers, RPC services, L2s)
3. Form initial hypotheses about partnership viability

## 2026-04-20 00:40 UTC

### Bootstrap Phase Completed
- **Literature Review:** Investigated Web3 affiliate programs and blockchain infrastructure providers
  - Source 1: UBUCH - Web3 affiliate programs overview
  - Source 2: Daily.dev - Developer tool affiliate programs (15 programs with 10-70% commissions)
  - Source 3: CompareNodes.com - 204 providers, 608 protocols, 2421 endpoints
- **Key Findings:**
  - Market is large: 204 providers, no clear leader in developer tools layer
  - Multiple partnership models: affiliate/referral (10-70%, 7-45d cookies), integration (20-30%, 60-90d), ecosystem grants (90-180d)
  - Privacy-first positioning is undererved by current infrastructure providers
  - Multi-provider complexity is a major pain point for developers

### Direction Decision: DEEPEN
- **Decision:** Move from bootstrap to inner loop
- **Reason:** Bootstrap confirmed market size and clear opportunity. Multiple partnership models identified. Ready to test hypotheses.

### Hypotheses Formed
Formulated 5 testable hypotheses prioritized by speed-to-revenue:

1. **H1: Affiliate Revenue Model** (Priority 0)
   - Test affiliate/referral partnerships for fastest path to revenue (30 days)
   - Prediction: 10-70% commissions, 7-45 day cookies, first revenue achievable

2. **H2: Integration Partnership Model** (Priority 1)
   - Test strategic integration partnerships with top 3-5 infrastructure providers
   - Prediction: 20-30% revenue share, 60-90 days setup

3. **H3: Privacy-First Positioning** (Priority 1)
   - Test privacy-first positioning as differentiator with privacy-conscious providers
   - Prediction: Premium positioning, exclusive partnerships possible

4. **H4: Ecosystem Grant Opportunities** (Priority 2)
   - Test L2 ecosystem grant programs (Base, Polygon, Arbitrum, Optimism)
   - Prediction: Non-dilutive funding + visibility, 90-180 days

5. **H5: Multi-Provider Management** (Priority 2)
   - Test multi-provider management tool as value-add for Formatho
   - Prediction: High-value differentiation, solves major pain point

### Next Action: Begin Inner Loop
- **Action:** Start H1 experiment - Affiliate Revenue Model
- **Protocol:** experiments/H1-affiliate-revenue-model/protocol.md
- **Execution:** Investigate affiliate programs for top 5 infrastructure providers, assess feasibility, create integration plan

## 2026-04-20 01:00 UTC

### H1 Experiment Completed: Affiliate Programs for Top 5 Providers
- **Experiment H1**: Affiliate Revenue Model testing top 5 infrastructure providers
- **Methods:** Web search using DuckDuckGo, Camoufox browser automation for verification
- **Results collected:**
  
  **1. Alchemy** (Feasibility: 2/10)
  - No affiliate program available
  - Alternative: Startup Program with credits, AWS credits, support
  - Not suitable for 30-day revenue goal
  
  **2. Chainstack** (Feasibility: 7/10)
  - $100 USDC per dedicated node referral
  - Instant payout in USDC
  - Sign up: https://chainstack.referral-factory.com/yIRjan
  - Target: High-value customers (dedicated node plans)
  
  **3. Ankr** (Feasibility: 8/10)
  - 50% of Liquid Staking rewards fees
  - Monthly payout via partner smart contract
  - Contact: staking@ankr.com
  - Example: 100k BNB staked → 0.005 BNB/day (at 1% TVL)
  
  **4. QuickNode** (Feasibility: 6/10)
  - PPS (Pay Per Sale) commission model
  - Platform: affiliates.quiknode.io
  - Multiple payment methods (PayPal, Bank Transfer, Bitcoin, etc.)
  - Exact commission rate requires signup
  
  **5. Tatum** (Feasibility: 9/10)
  - 15% recurring commission on referred payments
  - Platform: PromoteKit
  - Sign up: https://affiliates.tatum.io/
  - Highest feasibility for Formatho (recurring revenue, simple integration)

### Key Insights
- **Fastest path to revenue**: Tatum (15% recurring), Ankr (recurring), Chainstack (instant one-time)
- **Privacy opportunity**: No provider explicitly positions on privacy; Formatho can differentiate
- **Integration complexity**: Tatum/Chainstack (low), QuickNode/Ankr (medium), Alchemy (N/A)

### Direction Decision: DEEPEN
- **Decision**: H1 confirmed affiliate model viability; Tatum and Ankr offer best recurring revenue
- **Reason**: Strong evidence for 30-day revenue path through affiliate programs
- **Next**: Prepare H2 experiment (Integration Partnerships) for deeper partnerships

### Next Action: Update Research State
- Update research-state.yaml with H1 results
- Begin H2 experiment preparation
- Consider generating progress PDF report

