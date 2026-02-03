"use client";

import React from "react";
import { motion } from "framer-motion";
import { useSubscriptionPlans } from "@/hooks/use-subscription-queries";

interface PricingCardProps {
  tier: string;
  price: string | number;
  sub: string;
  features: string[];
  recommended: boolean;
  delay: number;
  buttonText: string;
}

const PricingCard = ({
  tier,
  price,
  sub,
  features,
  recommended,
  delay,
  buttonText,
  onSelect,
}: PricingCardProps & { onSelect: () => void }) => (
  <motion.div
    className={`relative p-6 sm:p-8 rounded-2xl sm:rounded-3xl transition-all duration-300 flex flex-col h-full group
            ${
              recommended
                ? "bg-slate-800/60 border border-slate-600 shadow-2xl scale-105 z-10 backdrop-blur-md"
                : "bg-slate-900/40 border border-slate-800 hover:bg-slate-800/60 hover:border-slate-600"
            }`}
    initial={{ opacity: 0, y: 50 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ duration: 0.6, delay }}
    whileHover={{ y: -5, scale: recommended ? 1.05 : 1.02 }}
  >
    {recommended && (
      <motion.div
        className="absolute -top-3 left-1/2 transform -translate-x-1/2 bg-blue-500 text-white px-4 py-1 rounded-full text-xs font-bold shadow-lg shadow-blue-500/30"
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: delay + 0.3, type: "spring", stiffness: 500 }}
      >
        Most Popular
      </motion.div>
    )}
    <div className="mb-8">
      <h3
        className={`text-lg font-bold mb-2 ${recommended ? "text-white" : "text-slate-200"}`}
      >
        {tier}
      </h3>
      <div className="flex items-baseline gap-1">
        {price === "Custom" ? (
          <span
            className={`text-3xl font-extrabold ${recommended ? "text-white" : "text-white"}`}
          >
            Custom Pricing
          </span>
        ) : (
          <>
            <span
              className={`text-4xl font-extrabold ${recommended ? "text-white" : "text-white"}`}
            >
              ${price}
            </span>
            <span className="text-slate-400 font-medium">/mo</span>
          </>
        )}
      </div>
      <p className="text-slate-500 text-sm mt-3">{sub}</p>
    </div>

    <div className="flex-1">
      <p className="text-sm font-semibold text-slate-300 mb-4">Includes:</p>
      <ul className="space-y-4 mb-8">
        {features.map((feature, idx) => (
          <motion.li
            key={idx}
            className="flex items-center gap-3 text-slate-400 group-hover:text-slate-300 transition-colors"
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ delay: delay + idx * 0.1 }}
          >
            <div
              className={`mt-0.5 w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 border ${recommended ? "border-blue-500 bg-blue-500/20 text-blue-400" : "border-slate-700 bg-slate-800 text-slate-400"}`}
            >
              <i className="fas fa-check text-[10px]"></i>
            </div>
            <span className="text-sm leading-tight">{feature}</span>
          </motion.li>
        ))}
      </ul>
    </div>

    <motion.button
      className={`w-full py-3.5 rounded-full font-bold text-sm transition-all ${
        recommended
          ? "bg-primary-600 hover:bg-primary-500 text-white shadow-[0_0_20px_rgba(37,99,235,0.3)]"
          : "bg-transparent border border-slate-600 text-white hover:bg-slate-800 hover:border-slate-500"
      }`}
      whileHover={{ scale: 1.02 }}
      whileTap={{ scale: 0.98 }}
      onClick={onSelect}
    >
      {buttonText}
    </motion.button>
  </motion.div>
);

export const Pricing = () => {
  // Use the shared TanStack Query hook for fetching subscription plans
  const { data: apiPlans = [], isLoading } = useSubscriptionPlans();

  // Transform API plans to match the expected format for PricingCard
  const plans = apiPlans.map((plan: any) => ({
    tier: plan.name,
    price: plan.slug === "ENTERPRISE" ? "Custom" : plan.priceMonthly.toString(),
    sub: plan.description,
    features: plan.features,
    recommended: plan.slug === "PRO_PLAN",
    buttonText:
      plan.slug === "ENTERPRISE" ? "Contact Sales" : `Choose ${plan.name}`,
    slug: plan.slug,
  }));

  const handlePlanSelect = (planSlug: string) => {
    if (planSlug === "ENTERPRISE") {
      // Enterprise requires contact
      window.open(
        "mailto:sales@liyali.com?subject=Enterprise Plan Inquiry",
        "_blank",
      );
    } else {
      // Redirect to register with plan parameter
      window.location.href = `/register?plan=${planSlug}&trial=true`;
    }
  };

  return (
    <section
      id="pricing"
      className="py-16 sm:py-20 lg:py-24 bg-[#050B14] relative overflow-hidden"
    >
      {/* Dark Blue Theme Blobs & Floating Math Operators */}
      <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none select-none">
        {/* Decorative Gradient Blobs */}
        <motion.div
          className="absolute top-[20%] left-[20%] w-[500px] h-[500px] bg-blue-600/10 rounded-full blur-[120px]"
          animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.6, 0.3] }}
          transition={{ duration: 8, repeat: Infinity }}
        />
        <motion.div
          className="absolute bottom-[10%] right-[10%] w-[400px] h-[400px] bg-indigo-900/20 rounded-full blur-[100px]"
          animate={{ scale: [1.2, 1, 1.2], opacity: [0.2, 0.5, 0.2] }}
          transition={{ duration: 10, repeat: Infinity, delay: 2 }}
        />

        {/* Floating 3D/Glass Math Operators */}
        <motion.div
          className="absolute top-[15%] left-[8%] text-8xl font-black text-blue-500/10 blur-[1px]"
          animate={{ y: [0, -20, 0], rotate: [0, 10, 0] }}
          transition={{ duration: 8, repeat: Infinity, ease: "easeInOut" }}
        >
          +
        </motion.div>
        <motion.div
          className="absolute bottom-[20%] right-[8%] text-9xl font-black text-indigo-400/10 blur-[2px]"
          animate={{ y: [0, 20, 0], rotate: [0, -15, 0] }}
          transition={{
            duration: 10,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 1,
          }}
        >
          ×
        </motion.div>
        <motion.div
          className="absolute top-[40%] right-[20%] text-7xl font-black text-sky-500/5 blur-[1px]"
          animate={{ y: [0, -15, 0], x: [0, 10, 0] }}
          transition={{
            duration: 12,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 2,
          }}
        >
          %
        </motion.div>
        <motion.div
          className="absolute bottom-[15%] left-[25%] text-8xl font-black text-blue-300/5 blur-[2px]"
          animate={{ y: [0, 25, 0], rotate: [0, 20, 0] }}
          transition={{
            duration: 9,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 0.5,
          }}
        >
          ÷
        </motion.div>
        <motion.div
          className="absolute top-[10%] right-[35%] text-6xl font-black text-primary-500/5"
          animate={{ scale: [1, 1.1, 1], opacity: [0.3, 0.7, 0.3] }}
          transition={{
            duration: 11,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 3,
          }}
        >
          =
        </motion.div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <motion.div
          className="text-center mb-12 sm:mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8 }}
        >
          <div className="inline-block px-4 py-1 rounded-full border border-slate-700 bg-slate-800/50 text-slate-300 text-xs font-bold mb-4 backdrop-blur-sm">
            <i className="fas fa-tag mr-2 text-primary-400"></i> Flexible Plans
          </div>
          <h2 className="text-2xl sm:text-3xl md:text-5xl font-extrabold text-white mb-4 sm:mb-6">
            The Right Plan for Every Business
          </h2>
          <p className="text-base sm:text-lg text-slate-400 max-w-2xl mx-auto">
            Choose the plan that fits your organization's size and needs.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 sm:gap-6 max-w-5xl mx-auto items-stretch">
          {isLoading
            ? // Loading skeleton
              Array.from({ length: 3 }).map((_, index) => (
                <div
                  key={index}
                  className="relative p-6 sm:p-8 rounded-2xl sm:rounded-3xl bg-slate-900/40 border border-slate-800 animate-pulse"
                >
                  <div className="mb-8">
                    <div className="h-6 w-24 bg-slate-700 rounded mb-2" />
                    <div className="h-10 w-32 bg-slate-700 rounded mb-3" />
                    <div className="h-4 w-40 bg-slate-800 rounded" />
                  </div>
                  <div className="space-y-4 mb-8">
                    {Array.from({ length: 5 }).map((_, i) => (
                      <div key={i} className="flex items-center gap-3">
                        <div className="w-5 h-5 rounded-full bg-slate-700" />
                        <div className="h-4 flex-1 bg-slate-800 rounded" />
                      </div>
                    ))}
                  </div>
                  <div className="h-12 w-full bg-slate-700 rounded-full" />
                </div>
              ))
            : plans.map((plan: any, index: number) => (
                <PricingCard
                  key={plan.tier}
                  tier={plan.tier}
                  price={plan.price}
                  sub={plan.sub}
                  features={plan.features}
                  recommended={plan.recommended}
                  delay={index * 0.1}
                  buttonText={plan.buttonText}
                  onSelect={() => handlePlanSelect(plan.slug)}
                />
              ))}
        </div>

        {/* Free Trial Banner */}
        <motion.div
          className="mt-12 max-w-4xl mx-auto"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className="relative rounded-3xl overflow-hidden bg-gradient-to-r from-blue-800/80 to-purple-800/80 border border-blue-700/50 p-1 md:p-2 backdrop-blur-md">
            <div className="absolute inset-0 bg-blue-500/5"></div>
            <div className="relative rounded-2xl bg-blue-900/20 px-6 py-6 md:px-12 md:flex items-center justify-between gap-6">
              <div className="text-center md:text-left">
                <h3 className="text-xl md:text-2xl font-bold text-white mb-2">
                  🎉 Start Your 14-Day Free Trial
                </h3>
                <p className="text-blue-200 text-sm">
                  Try all features risk-free. No credit card required to start
                  your trial.
                </p>
              </div>
              <motion.button
                className="mt-4 md:mt-0 w-full md:w-auto bg-white text-blue-900 px-8 py-3 rounded-full font-bold shadow-lg hover:bg-blue-50 transition-all"
                whileHover={{ y: -4, scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => (window.location.href = "/register?trial=true")}
              >
                Start Free Trial
              </motion.button>
            </div>
          </div>
        </motion.div>

        {/* Bottom Banner */}
        <motion.div
          className="mt-20 max-w-4xl mx-auto"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.3 }}
        >
          <div className="relative rounded-3xl overflow-hidden bg-gradient-to-r from-slate-800/80 to-slate-900/80 border border-slate-700/50 p-1 md:p-2 backdrop-blur-md">
            <div className="absolute inset-0 bg-blue-500/5"></div>
            <div className="relative rounded-2xl bg-blue-900/20 px-6 py-8 md:px-12 md:flex items-center justify-between gap-6">
              <div>
                <h3 className="text-xl md:text-2xl font-bold text-white mb-2">
                  Ready to Transform Your Operations?
                </h3>
                <p className="text-slate-400 text-sm">
                  Book a personalized demo to see how Liyali Suite can
                  streamline your procurement.
                </p>
              </div>
              <motion.button
                className="mt-4 md:mt-0 w-full md:w-auto bg-primary-600 hover:bg-primary-500 text-white px-8 py-3 rounded-full font-bold shadow-lg shadow-blue-500/20 transition-all"
                whileHover={{ y: -4, scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
              >
                Book Demo Now!
              </motion.button>
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
};
