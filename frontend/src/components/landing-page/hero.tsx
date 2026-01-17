"use client";

import React from "react";
import { motion } from "framer-motion";
import Link from "next/link";

export const Hero = () => {
  return (
    <section
      className="relative pt-24 pb-12 sm:pt-32 sm:pb-16 lg:pt-56 lg:pb-32 overflow-hidden bg-slate-50"
      aria-label="Hero section"
    >
      {/* Background Blobs */}
      <motion.div
        className="blob bg-blue-200/50 w-96 h-96 rounded-full absolute top-0 left-0 -translate-x-1/2 -translate-y-1/2 blur-3xl"
        animate={{
          scale: [1, 1.2, 1],
          rotate: [0, 180, 360],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: "linear",
        }}
      />
      <motion.div
        className="blob bg-indigo-200/50 w-80 h-80 rounded-full absolute bottom-0 right-0 translate-x-1/3 translate-y-1/3 blur-3xl"
        animate={{
          scale: [1.2, 1, 1.2],
          rotate: [360, 180, 0],
        }}
        transition={{
          duration: 25,
          repeat: Infinity,
          ease: "linear",
          delay: 2,
        }}
      />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
        <div className="grid lg:grid-cols-2 gap-8 lg:gap-12 items-center">
          <motion.div
            className="text-center lg:text-left"
            initial={{ opacity: 0, y: 50 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8, delay: 0.1 }}
          >
            <motion.div
              className="inline-block px-4 py-1.5 rounded-full bg-white border border-slate-200 text-primary-700 font-semibold text-sm mb-6 shadow-sm"
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.6, delay: 0.3 }}
            >
              <i
                className="fas fa-sparkles mr-2 text-primary-500"
                aria-hidden="true"
              ></i>
              The All-in-One Business Operating Platform
            </motion.div>
            <motion.h1
              className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-extrabold text-slate-900 leading-tight mb-4 sm:mb-6"
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.4 }}
            >
              Streamline operations with{" "}
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary-600 to-indigo-600">
                Liyali Suite
              </span>
            </motion.h1>
            <motion.p
              className="text-base sm:text-lg text-slate-600 mb-6 sm:mb-8 leading-relaxed max-w-2xl mx-auto lg:mx-0"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.5 }}
            >
              Enhance collaboration and drive growth with the unified platform
              designed for procurement, workflow automation, and modern team
              operations.
            </motion.p>
            <motion.div
              className="flex flex-col sm:flex-row items-center justify-center lg:justify-start gap-3 sm:gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8, delay: 0.6 }}
            >
              <Link href="/register">
                <motion.button
                  className="w-full sm:w-auto bg-primary-600 hover:bg-primary-700 text-white px-6 sm:px-8 py-3 sm:py-4 rounded-full font-bold text-base sm:text-lg transition-all shadow-xl shadow-primary-500/30 hover:shadow-primary-500/50 flex items-center justify-center gap-2"
                  whileHover={{ y: -4, scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  Start Free Trial{" "}
                  <i className="fas fa-arrow-right" aria-hidden="true"></i>
                </motion.button>
              </Link>
              <motion.button
                className="w-full sm:w-auto bg-white hover:bg-slate-50 text-slate-700 border border-slate-200 px-6 sm:px-8 py-3 sm:py-4 rounded-full font-bold text-base sm:text-lg transition-all hover:shadow-lg flex items-center justify-center gap-2"
                whileHover={{ y: -2, scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
              >
                <i
                  className="fas fa-play-circle text-primary-500 text-xl"
                  aria-hidden="true"
                ></i>{" "}
                Watch Demo
              </motion.button>
            </motion.div>
            <motion.div
              className="mt-8 flex items-center justify-center lg:justify-start gap-4 text-sm text-slate-500 font-medium"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.8, delay: 0.8 }}
            >
              <span className="flex items-center gap-1">
                <i className="fas fa-check-circle text-green-500"></i> No credit
                card required
              </span>
              <span className="flex items-center gap-1">
                <i className="fas fa-check-circle text-green-500"></i> 14-day
                free trial
              </span>
            </motion.div>
          </motion.div>

          <motion.div
            className="relative"
            initial={{ opacity: 0, x: 50 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.8, delay: 0.3 }}
          >
            <motion.div
              className="relative mx-auto rounded-2xl shadow-2xl bg-white border border-slate-200/60 overflow-hidden"
              whileHover={{ scale: 1.02, rotateY: 5 }}
              transition={{ type: "spring", stiffness: 300, damping: 30 }}
            >
              {/* Mock UI Header */}
              <div className="bg-slate-50 border-b border-slate-100 p-4 flex items-center gap-4">
                <div className="flex gap-2">
                  <div className="w-3 h-3 rounded-full bg-red-400"></div>
                  <div className="w-3 h-3 rounded-full bg-amber-400"></div>
                  <div className="w-3 h-3 rounded-full bg-green-400"></div>
                </div>
                <div className="flex-1 bg-white h-6 rounded-md shadow-sm"></div>
              </div>
              {/* Mock UI Body */}
              <div className="p-6 grid grid-cols-3 gap-6">
                <div className="col-span-1 space-y-4">
                  <motion.div
                    className="h-20 bg-primary-50 rounded-xl"
                    animate={{ scale: [1, 1.05, 1] }}
                    transition={{ duration: 2, repeat: Infinity }}
                  />
                  <div className="h-8 bg-slate-100 rounded-lg w-3/4"></div>
                  <div className="h-4 bg-slate-50 rounded w-full"></div>
                  <div className="h-4 bg-slate-50 rounded w-5/6"></div>
                </div>
                <div className="col-span-2 space-y-4">
                  <div className="flex justify-between">
                    <div className="h-10 w-32 bg-slate-100 rounded-lg"></div>
                    <motion.div
                      className="h-10 w-10 bg-primary-600 rounded-full shadow-lg shadow-primary-500/40"
                      animate={{ rotate: 360 }}
                      transition={{
                        duration: 3,
                        repeat: Infinity,
                        ease: "linear",
                      }}
                    />
                  </div>
                  <div className="h-40 bg-gradient-to-br from-slate-50 to-slate-100 rounded-xl border border-slate-100"></div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="h-24 bg-white shadow-sm border border-slate-100 rounded-xl"></div>
                    <div className="h-24 bg-white shadow-sm border border-slate-100 rounded-xl"></div>
                  </div>
                </div>
              </div>
            </motion.div>
            {/* Floating Badge */}
            <motion.div
              className="absolute -bottom-6 -left-6 bg-white p-4 rounded-2xl shadow-xl border border-slate-100 flex items-center gap-3"
              animate={{ y: [0, -10, 0] }}
              transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
            >
              <div className="bg-green-100 text-green-600 p-2 rounded-lg">
                <i className="fas fa-chart-line text-xl"></i>
              </div>
              <div>
                <p className="text-xs text-slate-500 font-semibold uppercase">
                  Efficiency
                </p>
                <p className="text-lg font-bold text-slate-900">+70%</p>
              </div>
            </motion.div>
          </motion.div>
        </div>
      </div>
    </section>
  );
};
