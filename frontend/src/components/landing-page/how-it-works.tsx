'use client';

import React from 'react';
import { motion } from 'framer-motion';

export const HowItWorks = () => {
    const steps = [
        {
            title: "Create Requisition",
            desc: "Staff submit requests with budgets, justifications, and attachments.",
            icon: "fa-file-invoice"
        },
        {
            title: "Automated Routing",
            desc: "System routes to HOD, Finance, and leadership based on your rules.",
            icon: "fa-network-wired"
        },
        {
            title: "Smart Procurement",
            desc: "RFQ management, vendor selection, and evaluation reports.",
            icon: "fa-shopping-cart"
        },
        {
            title: "Doc Automation",
            desc: "Auto-generate POs, GRNs, and payment vouchers with approval chains.",
            icon: "fa-file-contract"
        },
        {
            title: "Payment Execution",
            desc: "Integrated payment processing with bank/IFMIS connectivity.",
            icon: "fa-credit-card"
        }
    ];

    return (
        <section id="how-it-works" className="py-24 bg-slate-50 relative">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <motion.div 
                    className="text-center max-w-3xl mx-auto mb-16"
                    initial={{ opacity: 0, y: 30 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.8 }}
                >
                    <span className="text-primary-600 font-bold tracking-wider uppercase text-sm">Streamline Your Procurement</span>
                    <h2 className="text-3xl md:text-4xl font-extrabold text-slate-900 mt-2 mb-4">From Request to Payment in Minutes</h2>
                    <p className="text-lg text-slate-600">Liyali Suite transforms your cycle into a seamless, automated experience.</p>
                </motion.div>

                <div className="relative">
                     {/* Connecting Line (Desktop) */}
                    <motion.div 
                        className="hidden md:block absolute top-12 left-0 w-full h-1 bg-gradient-to-r from-primary-200 via-blue-200 to-indigo-200 rounded-full -z-10"
                        initial={{ scaleX: 0 }}
                        whileInView={{ scaleX: 1 }}
                        viewport={{ once: true }}
                        transition={{ duration: 1.5, delay: 0.5 }}
                    />

                    <div className="grid grid-cols-1 md:grid-cols-5 gap-8">
                        {steps.map((step, idx) => (
                            <motion.div 
                                key={idx}
                                className="relative flex flex-col items-center text-center group"
                                initial={{ opacity: 0, y: 50 }}
                                whileInView={{ opacity: 1, y: 0 }}
                                viewport={{ once: true }}
                                transition={{ duration: 0.6, delay: idx * 0.1 }}
                            >
                                <motion.div 
                                    className="w-24 h-24 bg-white rounded-full border-4 border-slate-50 shadow-lg flex items-center justify-center mb-6 transition-all duration-300"
                                    whileHover={{ 
                                        scale: 1.1, 
                                        borderColor: "rgb(59 130 246 / 0.3)",
                                        boxShadow: "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 10px 10px -5px rgb(0 0 0 / 0.04)"
                                    }}
                                >
                                    <i className={`fas ${step.icon} text-3xl text-primary-600`}></i>
                                </motion.div>
                                <h3 className="text-lg font-bold text-slate-900 mb-2">{step.title}</h3>
                                <p className="text-sm text-slate-600 leading-relaxed">{step.desc}</p>
                            </motion.div>
                        ))}
                    </div>
                </div>

                <motion.div 
                    className="mt-16 bg-white rounded-2xl p-8 border border-slate-200 shadow-sm flex flex-col md:flex-row items-center justify-between gap-6"
                    initial={{ opacity: 0, y: 30 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.8, delay: 0.3 }}
                >
                    <div className="flex items-start gap-4">
                        <motion.div 
                            className="bg-primary-100 p-3 rounded-lg text-primary-600"
                            whileHover={{ scale: 1.1, rotate: 5 }}
                        >
                            <i className="fas fa-database text-xl"></i>
                        </motion.div>
                        <div>
                            <h4 className="font-bold text-slate-900 mb-1">Flexible Workflow Engine</h4>
                            <p className="text-slate-600 text-sm">Database-driven workflows, condition-based routing, and offline capabilities.</p>
                        </div>
                    </div>
                    <div className="h-px w-full md:w-px md:h-12 bg-slate-200"></div>
                     <div className="flex items-start gap-4">
                        <motion.div 
                            className="bg-indigo-100 p-3 rounded-lg text-indigo-600"
                            whileHover={{ scale: 1.1, rotate: -5 }}
                        >
                            <i className="fas fa-lock text-xl"></i>
                        </motion.div>
                        <div>
                            <h4 className="font-bold text-slate-900 mb-1">Audit Ready</h4>
                            <p className="text-slate-600 text-sm">Complete document chain from request to payment proof for full transparency.</p>
                        </div>
                    </div>
                </motion.div>
            </div>
        </section>
    );
};