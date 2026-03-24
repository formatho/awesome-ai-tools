import { useState } from 'react';
import { motion } from 'framer-motion';

interface EmailCaptureFormProps {
  variant?: 'hero' | 'inline' | 'footer';
  source: string;
  ctaText?: string;
  placeholder?: string;
  className?: string;
}

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:18765';

export default function EmailCaptureForm({
  variant = 'inline',
  source,
  ctaText = 'Subscribe',
  placeholder = 'Enter your email',
  className = '',
}: EmailCaptureFormProps) {
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');
  const [message, setMessage] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email) {
      setStatus('error');
      setMessage('Please enter your email address');
      return;
    }

    setStatus('loading');

    try {
      const response = await fetch(`${API_BASE_URL}/api/newsletter/subscribe`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, source }),
      });

      const data = await response.json();

      if (response.ok) {
        setStatus('success');
        setMessage('🎉 Successfully subscribed! Check your inbox.');
        setEmail('');
      } else {
        setStatus('error');
        setMessage(data.error || 'Failed to subscribe. Please try again.');
      }
    } catch (error) {
      setStatus('error');
      setMessage('Network error. Please try again.');
    }
  };

  if (variant === 'hero') {
    return (
      <div className={`max-w-md mx-auto ${className}`}>
        <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-3">
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={placeholder}
            className="flex-1 px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
            disabled={status === 'loading'}
          />
          <motion.button
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            type="submit"
            disabled={status === 'loading'}
            className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {status === 'loading' ? 'Subscribing...' : ctaText}
          </motion.button>
        </form>
        {status !== 'idle' && (
          <motion.p
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className={`mt-3 text-sm ${
              status === 'success' ? 'text-green-600' : status === 'error' ? 'text-red-600' : 'text-gray-600'
            }`}
          >
            {message}
          </motion.p>
        )}
      </div>
    );
  }

  if (variant === 'footer') {
    return (
      <div className={className}>
        <form onSubmit={handleSubmit} className="flex flex-col gap-2">
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={placeholder}
            className="px-4 py-2 rounded border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all text-sm"
            disabled={status === 'loading'}
          />
          <button
            type="submit"
            disabled={status === 'loading'}
            className="btn-primary text-sm py-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {status === 'loading' ? 'Subscribing...' : ctaText}
          </button>
        </form>
        {status !== 'idle' && (
          <p
            className={`mt-2 text-xs ${
              status === 'success' ? 'text-green-600' : status === 'error' ? 'text-red-600' : 'text-gray-600'
            }`}
          >
            {message}
          </p>
        )}
      </div>
    );
  }

  // Default inline variant
  return (
    <div className={`card ${className}`}>
      <h3 className="text-xl font-semibold mb-3">Stay Updated</h3>
      <p className="text-gray-600 mb-4">Get the latest news and updates about Agent Orchestrator.</p>
      <form onSubmit={handleSubmit} className="flex flex-col gap-3">
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={placeholder}
          className="px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
          disabled={status === 'loading'}
        />
        <motion.button
          whileHover={{ scale: 1.01 }}
          whileTap={{ scale: 0.99 }}
          type="submit"
          disabled={status === 'loading'}
          className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {status === 'loading' ? 'Subscribing...' : ctaText}
        </motion.button>
      </form>
      {status !== 'idle' && (
        <motion.p
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className={`mt-3 text-sm ${
            status === 'success' ? 'text-green-600' : status === 'error' ? 'text-red-600' : 'text-gray-600'
          }`}
        >
          {message}
        </motion.p>
      )}
    </div>
  );
}
