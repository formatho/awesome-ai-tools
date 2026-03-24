import { useState } from 'react';
import { motion } from 'framer-motion';
import EmailCaptureForm from './components/EmailCaptureForm';
import TrustBadges from './components/TrustBadges';
import './App.css';

function App() {
  const [activeFeature, setActiveFeature] = useState(0);

  const features = [
    {
      title: 'Multi-Agent Orchestration',
      description: 'Coordinate multiple AI agents working together seamlessly',
      icon: '🤖',
    },
    {
      title: 'Local-First Privacy',
      description: 'Your data stays on your machine. No cloud dependency.',
      icon: '🔒',
    },
    {
      title: 'Smart Task Queue',
      description: 'Priority-based task management with intelligent routing',
      icon: '📋',
    },
    {
      title: 'Cron Scheduling',
      description: 'Automate agent tasks with flexible scheduling',
      icon: '⏰',
    },
    {
      title: 'Skill System',
      description: 'Extensible capabilities with granular permissions',
      icon: '⚡',
    },
    {
      title: 'Multi-LLM Support',
      description: 'Works with OpenAI, Anthropic, Ollama, and more',
      icon: '🧠',
    },
  ];

  const steps = [
    {
      step: '01',
      title: 'Configure Agents',
      description: 'Set up your AI agents with specific roles and skills',
    },
    {
      step: '02',
      title: 'Define Tasks',
      description: 'Create and prioritize tasks for your agent workforce',
    },
    {
      step: '03',
      title: 'Watch Magic Happen',
      description: 'Agents collaborate to complete tasks automatically',
    },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-50 to-white">
      {/* Header */}
      <header className="fixed top-0 left-0 right-0 bg-white/80 backdrop-blur-md z-50 border-b border-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-2xl">🏗️</span>
            <span className="text-xl font-bold text-gray-900">Agent Orchestrator</span>
          </div>
          <nav className="hidden md:flex items-center gap-8">
            <a href="#features" className="text-gray-600 hover:text-gray-900 transition-colors">
              Features
            </a>
            <a href="#how-it-works" className="text-gray-600 hover:text-gray-900 transition-colors">
              How It Works
            </a>
            <a href="#pricing" className="text-gray-600 hover:text-gray-900 transition-colors">
              Pricing
            </a>
            <button className="btn-primary">Get Started Free</button>
          </nav>
        </div>
      </header>

      {/* Hero Section */}
      <section className="pt-32 pb-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto text-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <h1 className="text-5xl md:text-6xl font-bold text-gray-900 mb-6">
              Your AI Agent
              <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-purple-600">
                Workforce, Orchestrated
              </span>
            </h1>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto mb-8">
              Coordinate multiple AI agents to automate complex workflows. Local-first, privacy-focused,
              and enterprise-ready. No cloud required.
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2, duration: 0.5 }}
          >
            <EmailCaptureForm
              variant="hero"
              source="homepage-hero"
              ctaText="Join Early Access"
              placeholder="Enter your email for early access"
              className="mb-4"
            />
            <p className="text-sm text-gray-500 mb-6">
              🎉 Free during beta • No credit card required • Cancel anytime
            </p>
          </motion.div>

          {/* Trust Badges - Hero Variant */}
          <TrustBadges variant="hero" />
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 px-4 sm:px-6 lg:px-8 bg-gray-50">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">
              Everything You Need to Orchestrate AI Agents
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              A complete toolkit for managing, scheduling, and coordinating your AI workforce
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <motion.div
                key={index}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1, duration: 0.5 }}
                viewport={{ once: true }}
                className="bg-white rounded-xl p-6 shadow-sm border border-gray-100 hover:shadow-md transition-shadow"
              >
                <div className="text-4xl mb-4">{feature.icon}</div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">{feature.title}</h3>
                <p className="text-gray-600">{feature.description}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section id="how-it-works" className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">How It Works</h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Get started in minutes with our simple three-step process
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {steps.map((step, index) => (
              <motion.div
                key={index}
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.2, duration: 0.5 }}
                viewport={{ once: true }}
                className="text-center"
              >
                <div className="text-6xl font-bold text-gray-200 mb-4">{step.step}</div>
                <h3 className="text-2xl font-semibold text-gray-900 mb-3">{step.title}</h3>
                <p className="text-gray-600">{step.description}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Trust Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-r from-blue-600 to-purple-600 text-white">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl font-bold mb-6">
            Why Developers Trust Agent Orchestrator
          </h2>
          <p className="text-xl mb-8 opacity-90">
            Built by developers, for developers. We understand the importance of privacy, reliability, and control.
          </p>
          <div className="grid md:grid-cols-3 gap-8">
            <div>
              <div className="text-5xl font-bold mb-2">100%</div>
              <div className="text-lg opacity-90">Local-First</div>
            </div>
            <div>
              <div className="text-5xl font-bold mb-2">0</div>
              <div className="text-lg opacity-90">Data Sent to Cloud</div>
            </div>
            <div>
              <div className="text-5xl font-bold mb-2">∞</div>
              <div className="text-lg opacity-90">Possibilities</div>
            </div>
          </div>
        </div>
      </section>

      {/* Bottom CTA Section */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-4xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            viewport={{ once: true }}
            className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8 md:p-12"
          >
            <div className="text-center mb-8">
              <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
                Ready to Orchestrate Your AI Agents?
              </h2>
              <p className="text-xl text-gray-600">
                Join the waitlist for early access and be the first to try Agent Orchestrator
              </p>
            </div>

            <EmailCaptureForm
              variant="hero"
              source="homepage-bottom"
              ctaText="Get Early Access"
              placeholder="your@email.com"
              className="mb-6"
            />

            {/* Trust Badges - Inline Variant */}
            <TrustBadges variant="inline" />
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-16 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            <div>
              <div className="flex items-center gap-2 mb-4">
                <span className="text-2xl">🏗️</span>
                <span className="text-xl font-bold">Agent Orchestrator</span>
              </div>
              <p className="text-gray-400">
                Your AI agent workforce, orchestrated with privacy and control.
              </p>
            </div>

            <div>
              <h4 className="font-semibold mb-4">Product</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#features" className="hover:text-white transition-colors">Features</a></li>
                <li><a href="#pricing" className="hover:text-white transition-colors">Pricing</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Documentation</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Changelog</a></li>
              </ul>
            </div>

            <div>
              <h4 className="font-semibold mb-4">Company</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#" className="hover:text-white transition-colors">About</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Blog</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Careers</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Contact</a></li>
              </ul>
            </div>

            <div>
              <h4 className="font-semibold mb-4">Stay Updated</h4>
              <EmailCaptureForm
                variant="footer"
                source="footer"
                ctaText="Subscribe"
                placeholder="your@email.com"
              />
            </div>
          </div>

          <div className="border-t border-gray-800 pt-8">
            <div className="flex flex-col md:flex-row items-center justify-between gap-4">
              <div className="text-gray-400 text-sm">
                © 2026 Formatho. All rights reserved.
              </div>
              
              {/* Trust Badges - Footer Variant */}
              <TrustBadges variant="footer" />
              
              <div className="flex items-center gap-4 text-gray-400 text-sm">
                <a href="#" className="hover:text-white transition-colors">Privacy Policy</a>
                <a href="#" className="hover:text-white transition-colors">Terms of Service</a>
              </div>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;
