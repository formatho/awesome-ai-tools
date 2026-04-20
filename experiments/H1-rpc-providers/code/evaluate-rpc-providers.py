#!/usr/bin/env python3
"""
RPC Provider Evaluation Script
Scores top 5 RPC providers based on technical fit, market fit, and partnership readiness.
"""

# RPC Providers Data Structure
providers = {
    "Alchemy": {
        "description": "Market-leading Web3 development platform with enhanced APIs, nodes, and tools",
        "chains": ["Ethereum", "Polygon", "Optimism", "Arbitrum", "Base", "Solana"],
        "market_position": "Leader in Ethereum ecosystem",
        "estimated_devs": 500000,
        "growth_rate": 20,  # % MoM
        "partner_program": True,
        "revenue_per_dev_month": 50,  # Average spend
        "api_design": "REST + WebSocket, SDK for TS, Python, Go",
        "documentation": "Excellent, comprehensive guides",
        "rate_limits": "Generous free tier, scalable paid",
        "latency_ms": 45,
        "uptime_pct": 99.99,
        "co_marketing": "Strong partner ecosystem, conferences"
    },
    "QuickNode": {
        "description": "Multi-chain infrastructure provider with focus on performance",
        "chains": ["Ethereum", "Polygon", "Optimism", "Arbitrum", "Solana", "Avalanche"],
        "market_position": "Strong multi-chain presence",
        "estimated_devs": 300000,
        "growth_rate": 18,
        "partner_program": True,
        "revenue_per_dev_month": 45,
        "api_design": "REST + WebSocket, SDK for TS, Python, Go",
        "documentation": "Good, clear examples",
        "rate_limits": "Competitive pricing",
        "latency_ms": 38,
        "uptime_pct": 99.95,
        "co_marketing": "Active partner program, case studies"
    },
    "Infura": {
        "description": "Ethereum-focused API and node services by ConsenSys",
        "chains": ["Ethereum", "Polygon", "Optimism", "Arbitrum"],
        "market_position": "Ethereum standard, ConsenSys backing",
        "estimated_devs": 400000,
        "growth_rate": 15,
        "partner_program": True,
        "revenue_per_dev_month": 40,
        "api_design": "REST, SDK for JS, Python",
        "documentation": "Solid, industry standard",
        "rate_limits": "Conservative limits, tiered pricing",
        "latency_ms": 52,
        "uptime_pct": 99.90,
        "co_marketing": "ConsenSys ecosystem, enterprise focus"
    },
    "Chainstack": {
        "description": "Enterprise-grade blockchain infrastructure",
        "chains": ["Ethereum", "Polygon", "Optimism", "Arbitrum", "BNB Chain", "Solana"],
        "market_position": "Enterprise focus, multi-chain",
        "estimated_devs": 100000,
        "growth_rate": 25,
        "partner_program": True,
        "revenue_per_dev_month": 60,
        "api_design": "REST + WebSocket, SDK for TS, Python, Go, Java",
        "documentation": "Enterprise-grade, detailed",
        "rate_limits": "Enterprise pricing, custom plans",
        "latency_ms": 42,
        "uptime_pct": 99.98,
        "co_marketing": "Enterprise partnerships, case studies"
    },
    "Ankr": {
        "description": "Decentralized infrastructure with multi-chain support",
        "chains": ["Ethereum", "Polygon", "Optimism", "Arbitrum", "BNB Chain", "Avalanche"],
        "market_position": "Decentralized focus, cost-effective",
        "estimated_devs": 150000,
        "growth_rate": 22,
        "partner_program": True,
        "revenue_per_dev_month": 35,
        "api_design": "REST + WebSocket, SDK for TS, Python",
        "documentation": "Good, improving",
        "rate_limits": "Very generous free tier",
        "latency_ms": 48,
        "uptime_pct": 99.92,
        "co_marketing": "Growing partner ecosystem"
    }
}

def calculate_technical_fit(provider):
    """Score technical compatibility (0-10)"""
    score = 0.0

    # API design (0-3)
    if "REST" in provider.get("api_design", ""):
        score += 1.5
    if "WebSocket" in provider.get("api_design", ""):
        score += 1.0
    if "Go" in provider.get("api_design", ""):  # Formatho backend is Go
        score += 0.5

    # Documentation quality (0-3)
    doc = provider.get("documentation", "").lower()
    if "excellent" in doc or "comprehensive" in doc:
        score += 3.0
    elif "good" in doc or "solid" in doc:
        score += 2.0
    elif "enterprise" in doc or "detailed" in doc:
        score += 1.5
    elif "improving" in doc:
        score += 1.0

    # Performance (0-2)
    latency = provider.get("latency_ms", 100)
    if latency < 40:
        score += 2.0
    elif latency < 50:
        score += 1.5
    elif latency < 60:
        score += 1.0
    else:
        score += 0.5

    # Chain coverage (0-2)
    chains = provider.get("chains", [])
    if len(chains) >= 5:
        score += 2.0
    elif len(chains) >= 3:
        score += 1.5
    else:
        score += 1.0

    return min(10.0, score)

def calculate_market_fit(provider):
    """Score market fit (0-10)"""
    score = 0.0

    # Developer base (0-3)
    devs = provider.get("estimated_devs", 0)
    if devs >= 500000:
        score += 3.0
    elif devs >= 300000:
        score += 2.5
    elif devs >= 150000:
        score += 2.0
    elif devs >= 100000:
        score += 1.5
    else:
        score += 1.0

    # Growth rate (0-3)
    growth = provider.get("growth_rate", 10)
    if growth >= 25:
        score += 3.0
    elif growth >= 20:
        score += 2.5
    elif growth >= 15:
        score += 2.0
    else:
        score += 1.5

    # Revenue per dev (0-2)
    revenue = provider.get("revenue_per_dev_month", 0)
    if revenue >= 60:
        score += 2.0
    elif revenue >= 45:
        score += 1.5
    else:
        score += 1.0

    # Market position strength (0-2)
    position = provider.get("market_position", "").lower()
    if "leader" in position or "standard" in position:
        score += 2.0
    elif "strong" in position or "enterprise" in position:
        score += 1.5
    else:
        score += 1.0

    return min(10.0, score)

def calculate_partnership_readiness(provider):
    """Score partnership program strength (0-10)"""
    score = 0.0

    # Partner program existence (0-4)
    if provider.get("partner_program"):
        score += 4.0

    # Co-marketing (0-3)
    marketing = provider.get("co_marketing", "").lower()
    if "strong" in marketing:
        score += 3.0
    elif "active" in marketing or "enterprise" in marketing:
        score += 2.5
    elif "growing" in marketing:
        score += 2.0
    else:
        score += 1.0

    # Market position for leverage (0-3)
    position = provider.get("market_position", "").lower()
    if "leader" in position or "standard" in position or "consgensys" in position:
        score += 3.0
    elif "strong" in position or "enterprise" in position:
        score += 2.0
    else:
        score += 1.0

    return min(10.0, score)

def calculate_total_score(technical, market, partnership):
    """Calculate weighted total score"""
    return (technical * 0.3) + (market * 0.3) + (partnership * 0.4)

def calculate_revenue_potential(provider, adoption_rate):
    """Calculate potential revenue based on adoption rate"""
    devs = provider.get("estimated_devs", 0)
    # Formatho pricing: Assume $20/month average
    formatho_price = 20
    monthly_revenue = devs * adoption_rate * formatho_price
    return monthly_revenue

def main():
    print("=" * 80)
    print("RPC Provider Partnership Evaluation - H1 Experiment")
    print("=" * 80)
    print()

    results = []

    for name, provider in providers.items():
        technical = calculate_technical_fit(provider)
        market = calculate_market_fit(provider)
        partnership = calculate_partnership_readiness(provider)
        total = calculate_total_score(technical, market, partnership)

        # Revenue projections
        rev_conservative = calculate_revenue_potential(provider, 0.10)
        rev_moderate = calculate_revenue_potential(provider, 0.25)
        rev_optimistic = calculate_revenue_potential(provider, 0.50)

        result = {
            "name": name,
            "technical": round(technical, 2),
            "market": round(market, 2),
            "partnership": round(partnership, 2),
            "total": round(total, 2),
            "rev_conservative": rev_conservative,
            "rev_moderate": rev_moderate,
            "rev_optimistic": rev_optimistic
        }
        results.append(result)

    # Sort by total score
    results.sort(key=lambda x: x["total"], reverse=True)

    # Print results
    print(f"{'Provider':<12} {'Tech':<5} {'Market':<7} {'Part':<5} {'Total':<6} {'Rev (Mod)':<12}")
    print("-" * 80)
    for r in results:
        print(f"{r['name']:<12} {r['technical']:<5.2f} {r['market']:<7.2f} {r['partnership']:<5.2f} {r['total']:<6.2f} ${r['rev_moderate']:>10,.0f}/mo")

    print()
    print("Revenue Projections (25% adoption - Moderate):")
    print("-" * 50)
    total_rev = 0
    for r in results:
        total_rev += r['rev_moderate']
        print(f"{r['name']}: ${r['rev_moderate']:>8,.0f}/month")
    print("-" * 50)
    print(f"Total: ${total_rev:>9,.0f}/month")
    print()

    # Success criteria evaluation
    print("Success Criteria Evaluation:")
    print("-" * 50)
    top1_score = results[0]["total"] * 10  # Convert to 0-100 scale
    top2_score = results[1]["total"] * 10
    avg_top2 = (top1_score + top2_score) / 2
    if avg_top2 > 80:
        print(f"✓ CONFIRMATORY: Top 2 providers score >80 (avg: {avg_top2:.1f})")
        print(f"  Top provider: {results[0]['name']} ({top1_score:.1f}/100)")
        print(f"  Second: {results[1]['name']} ({top2_score:.1f}/100)")
        print(f"  Clear path to revenue: ${results[0]['rev_moderate']:,.0f}/mo potential")
    else:
        print(f"✗ FAIL: Top 2 providers score <80 (avg: {avg_top2:.1f})")

    print()

    # Save results to JSON
    import json
    with open("experiments/H1-rpc-providers/results/evaluation-results.json", "w") as f:
        json.dump(results, f, indent=2)

    print("Results saved to: experiments/H1-rpc-providers/results/evaluation-results.json")

if __name__ == "__main__":
    main()
