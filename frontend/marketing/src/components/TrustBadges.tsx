import { motion } from 'framer-motion';

interface TrustBadgesProps {
  variant?: 'hero' | 'inline' | 'footer';
  className?: string;
}

const badges = [
  {
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
      </svg>
    ),
    label: 'SOC2 Compliant',
    description: 'Enterprise-grade security',
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    label: 'GDPR Ready',
    description: 'Full data compliance',
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
      </svg>
    ),
    label: '256-bit SSL',
    description: 'Encrypted connections',
  },
  {
    icon: (
      <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    label: '99.9% Uptime',
    description: 'Reliable service',
  },
];

export default function TrustBadges({ variant = 'inline', className = '' }: TrustBadgesProps) {
  if (variant === 'footer') {
    return (
      <div className={`flex flex-wrap justify-center gap-4 ${className}`}>
        {badges.map((badge, index) => (
          <div
            key={index}
            className="flex items-center gap-2 text-gray-400 text-sm"
          >
            <span className="text-green-500">{badge.icon}</span>
            <span>{badge.label}</span>
          </div>
        ))}
      </div>
    );
  }

  if (variant === 'hero') {
    return (
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.5, duration: 0.5 }}
        className={`flex flex-wrap justify-center gap-6 mt-8 ${className}`}
      >
        {badges.map((badge, index) => (
          <motion.div
            key={index}
            whileHover={{ scale: 1.05 }}
            className="flex items-center gap-2 bg-white/80 backdrop-blur-sm px-4 py-2 rounded-lg shadow-sm border border-gray-100"
          >
            <span className="text-green-600">{badge.icon}</span>
            <div className="flex flex-col">
              <span className="text-sm font-medium text-gray-900">{badge.label}</span>
              <span className="text-xs text-gray-500">{badge.description}</span>
            </div>
          </motion.div>
        ))}
      </motion.div>
    );
  }

  // Default inline variant
  return (
    <div className={`bg-gradient-to-r from-green-50 to-blue-50 rounded-xl p-6 ${className}`}>
      <div className="flex items-center justify-center gap-2 mb-4">
        <svg className="w-5 h-5 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
        <h3 className="text-lg font-semibold text-gray-900">Privacy-First, Enterprise-Ready</h3>
      </div>
      <p className="text-center text-gray-600 mb-6">
        Your data never leaves your machine. Built with security and privacy at the core.
      </p>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {badges.map((badge, index) => (
          <motion.div
            key={index}
            whileHover={{ scale: 1.05, y: -2 }}
            className="flex flex-col items-center text-center bg-white rounded-lg p-4 shadow-sm border border-gray-100"
          >
            <div className="text-green-600 mb-2">{badge.icon}</div>
            <span className="text-sm font-medium text-gray-900">{badge.label}</span>
            <span className="text-xs text-gray-500 mt-1">{badge.description}</span>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
