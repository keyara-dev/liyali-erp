"use client";

import { motion } from "framer-motion";

interface FeatureCardProps {
  icon: string;
  title: string;
  description: string;
  delay: number;
}

const FeatureCard = ({ icon, title, description, delay }: FeatureCardProps) => (
  <motion.div
    className="p-6 sm:p-8 rounded-2xl sm:rounded-3xl transition-all duration-300 group border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 shadow-lg hover:shadow-2xl"
    initial={{ opacity: 0, y: 50 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ duration: 0.6, delay }}
    whileHover={{ y: -8, scale: 1.02 }}
  >
    <motion.div
      className="w-14 h-14 bg-gradient-to-br from-primary-50 to-indigo-50 dark:from-primary-900/50 dark:to-indigo-900/50 rounded-2xl flex items-center justify-center text-primary-600 dark:text-primary-400 text-2xl mb-6 shadow-sm border border-primary-100 dark:border-primary-800"
      whileHover={{ scale: 1.1, rotate: 5 }}
      transition={{ type: "spring", stiffness: 400, damping: 10 }}
    >
      <i className={`fas ${icon}`}></i>
    </motion.div>
    <h3 className="text-lg sm:text-xl font-bold text-slate-900 dark:text-white mb-3">
      {title}
    </h3>
    <p className="text-sm sm:text-base text-slate-600 dark:text-slate-300 leading-relaxed">
      {description}
    </p>
  </motion.div>
);

export const Features = () => {
  const features = [
    {
      icon: "fa-layer-group",
      title: "Unified Dashboard",
      description:
        "All metrics in one customizable workspace. Track what matters most to your business.",
    },
    {
      icon: "fa-users",
      title: "Team Collaboration",
      description:
        "Real-time workflow approvals, commenting, and document sharing to keep teams aligned.",
    },
    {
      icon: "fa-bolt",
      title: "Lightning Fast",
      description:
        "Next-gen architecture ensuring zero latency and instant updates across all devices.",
    },
    {
      icon: "fa-chart-pie",
      title: "Advanced Analytics",
      description:
        "AI-powered insights for procurement, spending trends, and beautiful exportable reports.",
    },
    {
      icon: "fa-shield-alt",
      title: "Bank-Grade Security",
      description:
        "Enterprise-level encryption and compliance standards to keep your sensitive data secure.",
    },
    {
      icon: "fa-robot",
      title: "Smart Automation",
      description:
        "End-to-end procure-to-pay workflow automation to repetitive tasks and save hours.",
    },
  ];

  return (
    <section
      id="features"
      className="py-16 sm:py-20 lg:py-24 bg-white dark:bg-slate-900 relative transition-colors duration-300"
      aria-label="Product features"
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div
          className="text-center max-w-3xl mx-auto mb-12 sm:mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8 }}
        >
          <span className="text-primary-600 dark:text-primary-400 font-bold tracking-wider uppercase text-sm">
            Why Liyali Suite?
          </span>
          <h2 className="text-2xl sm:text-3xl md:text-4xl font-extrabold text-slate-900 dark:text-white mt-2 mb-4">
            Everything You Need to Scale
          </h2>
          <p className="text-base sm:text-lg text-slate-600 dark:text-slate-300">
            Powerful tools integrated into one seamless platform built for
            procurement, workflow automation, and team collaboration.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
          {features.map((feature, index) => (
            <FeatureCard
              key={feature.title}
              icon={feature.icon}
              title={feature.title}
              description={feature.description}
              delay={index * 0.1}
            />
          ))}
        </div>
      </div>
    </section>
  );
};
