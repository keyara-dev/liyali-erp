'use client';

import React from 'react';
import { motion } from 'framer-motion';

export const About = () => {
    return (
        <section id="about" className="py-24 bg-white relative overflow-hidden">
            {/* Background Decorations */}
            <motion.div 
                className="absolute top-0 right-0 -translate-y-1/2 translate-x-1/4 w-96 h-96 bg-blue-100/50 rounded-full blur-3xl -z-10"
                animate={{ scale: [1, 1.1, 1], rotate: [0, 90, 180] }}
                transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
            />
            <motion.div 
                className="absolute bottom-0 left-0 translate-y-1/2 -translate-x-1/4 w-80 h-80 bg-indigo-100/50 rounded-full blur-3xl -z-10"
                animate={{ scale: [1.1, 1, 1.1], rotate: [180, 270, 360] }}
                transition={{ duration: 25, repeat: Infinity, ease: "linear" }}
            />

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
                <div className="grid lg:grid-cols-2 gap-16 items-center mb-20">
                    <motion.div
                        initial={{ opacity: 0, x: -50 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.8 }}
                    >
                        <span className="text-primary-600 font-bold tracking-wider uppercase text-sm">About Liyali Suite</span>
                        <h2 className="text-3xl md:text-4xl font-extrabold text-slate-900 mt-2 mb-6">Built for Modern Business Operations</h2>
                        <p className="text-lg text-slate-600 leading-relaxed mb-6">
                            We understand that every minute spent on manual processes is a minute stolen from strategic growth. That's why we've engineered Liyali Suite to be the operational backbone that scales with your ambitions.
                        </p>
                        <p className="text-slate-600 leading-relaxed mb-8">
                            From startups to enterprise organizations, our platform adapts to your unique workflows while maintaining the security, compliance, and performance standards that modern businesses demand.
                        </p>
                        
                        <div className="grid grid-cols-2 gap-6">
                            <motion.div 
                                className="text-center p-4"
                                whileHover={{ scale: 1.05 }}
                                transition={{ type: "spring", stiffness: 300 }}
                            >
                                <div className="text-3xl font-bold text-primary-600 mb-2">500+</div>
                                <div className="text-sm text-slate-600 font-medium">Organizations Trust Us</div>
                            </motion.div>
                            <motion.div 
                                className="text-center p-4"
                                whileHover={{ scale: 1.05 }}
                                transition={{ type: "spring", stiffness: 300 }}
                            >
                                <div className="text-3xl font-bold text-primary-600 mb-2">99.9%</div>
                                <div className="text-sm text-slate-600 font-medium">Uptime Guarantee</div>
                            </motion.div>
                        </div>
                    </motion.div>

                    <motion.div
                        initial={{ opacity: 0, x: 50 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.8, delay: 0.2 }}
                    >
                        <div className="aspect-[4/3] relative rounded-2xl bg-gradient-to-br from-slate-50 to-white border border-slate-200 shadow-2xl overflow-hidden flex items-center justify-center p-8 group">
                            <div className="absolute inset-0 bg-[radial-gradient(#e2e8f0_1px,transparent_1px)] [background-size:16px_16px] opacity-50"></div>
                            
                            {/* Animated Abstract UI Elements */}
                            <div className="relative w-full max-w-sm">
                                <motion.div 
                                    className="absolute -top-12 -right-12 w-24 h-24 bg-blue-500/10 rounded-full blur-xl"
                                    animate={{ scale: [1, 1.2, 1], opacity: [0.3, 0.6, 0.3] }}
                                    transition={{ duration: 4, repeat: Infinity }}
                                />
                                
                                <motion.div 
                                    className="bg-white p-6 rounded-xl shadow-lg border border-slate-100 mb-4"
                                    initial={{ x: 20 }}
                                    whileInView={{ x: 0 }}
                                    whileHover={{ x: -10 }}
                                    transition={{ duration: 0.7 }}
                                >
                                    <div className="flex items-center gap-4 mb-3">
                                        <motion.div 
                                            className="w-10 h-10 rounded-lg bg-green-100 flex items-center justify-center text-green-600"
                                            animate={{ rotate: [0, 360] }}
                                            transition={{ duration: 3, repeat: Infinity, ease: "linear" }}
                                        >
                                            <i className="fas fa-check"></i>
                                        </motion.div>
                                        <div>
                                            <div className="h-2 w-24 bg-slate-200 rounded mb-1"></div>
                                            <div className="h-2 w-16 bg-slate-100 rounded"></div>
                                        </div>
                                    </div>
                                    <div className="h-1 w-full bg-slate-100 rounded overflow-hidden">
                                        <motion.div 
                                            className="h-full bg-green-500"
                                            initial={{ width: "0%" }}
                                            whileInView={{ width: "75%" }}
                                            transition={{ duration: 2, delay: 0.5 }}
                                        />
                                    </div>
                                </motion.div>

                                <motion.div 
                                    className="bg-white p-6 rounded-xl shadow-[0_20px_50px_rgba(0,0,0,0.1)] border border-slate-100 relative z-10"
                                    whileHover={{ scale: 1.05, rotateY: 5 }}
                                    transition={{ duration: 0.5 }}
                                >
                                    <div className="flex justify-between items-center mb-4">
                                        <h4 className="font-bold text-slate-800">Operational Efficiency</h4>
                                        <motion.span 
                                            className="text-green-500 text-sm font-bold"
                                            animate={{ scale: [1, 1.1, 1] }}
                                            transition={{ duration: 2, repeat: Infinity }}
                                        >
                                            +128%
                                        </motion.span>
                                    </div>
                                    <div className="space-y-2">
                                        <motion.div 
                                            className="h-12 bg-slate-50 rounded-lg border border-slate-100 flex items-center px-3"
                                            whileHover={{ backgroundColor: "rgb(248 250 252)" }}
                                        >
                                            <div className="w-2 h-2 rounded-full bg-blue-500 mr-2"></div>
                                            <div className="h-2 w-20 bg-slate-200 rounded"></div>
                                        </motion.div>
                                        <motion.div 
                                            className="h-12 bg-slate-50 rounded-lg border border-slate-100 flex items-center px-3"
                                            whileHover={{ backgroundColor: "rgb(248 250 252)" }}
                                        >
                                            <div className="w-2 h-2 rounded-full bg-indigo-500 mr-2"></div>
                                            <div className="h-2 w-20 bg-slate-200 rounded"></div>
                                        </motion.div>
                                    </div>
                                </motion.div>
                                
                                <motion.div 
                                    className="bg-white p-4 rounded-xl shadow-lg border border-slate-100 mt-4 w-3/4 ml-auto"
                                    initial={{ x: 10 }}
                                    whileInView={{ x: 0 }}
                                    whileHover={{ x: 20 }}
                                    transition={{ duration: 0.7, delay: 0.1 }}
                                >
                                     <div className="flex items-center gap-3">
                                        <motion.div 
                                            className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 text-xs"
                                            whileHover={{ rotate: 360 }}
                                            transition={{ duration: 0.5 }}
                                        >
                                            <i className="fas fa-user"></i>
                                        </motion.div>
                                        <div className="h-2 w-20 bg-slate-200 rounded"></div>
                                     </div>
                                </motion.div>
                            </div>
                        </div>
                    </motion.div>
                </div>

                <div className="grid md:grid-cols-2 gap-8">
                    {/* Mission Card */}
                    <motion.div 
                        className="bg-slate-50 p-10 rounded-3xl border border-slate-100 hover:border-blue-200 hover:shadow-xl hover:shadow-blue-500/5 transition-all duration-300 group"
                        initial={{ opacity: 0, y: 50 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.6, delay: 0.3 }}
                        whileHover={{ y: -5 }}
                    >
                        <motion.div 
                            className="w-16 h-16 bg-white rounded-2xl flex items-center justify-center text-3xl mb-8 shadow-sm border border-slate-100"
                            whileHover={{ scale: 1.1, rotate: 3 }}
                            transition={{ type: "spring", stiffness: 400, damping: 10 }}
                        >
                            <i className="fas fa-bullseye text-transparent bg-clip-text bg-gradient-to-br from-blue-500 to-blue-700"></i>
                        </motion.div>
                        <h3 className="text-2xl font-bold text-slate-900 mb-4">Our Mission</h3>
                        <p className="text-slate-600 text-lg leading-relaxed">
                            To dismantle operational silos and empower teams with a unified platform that makes procurement, approval, and execution effortless. We exist to give businesses back their most valuable asset: time.
                        </p>
                    </motion.div>

                    {/* Vision Card */}
                    <motion.div 
                        className="bg-slate-50 p-10 rounded-3xl border border-slate-100 hover:border-indigo-200 hover:shadow-xl hover:shadow-indigo-500/5 transition-all duration-300 group"
                        initial={{ opacity: 0, y: 50 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.6, delay: 0.4 }}
                        whileHover={{ y: -5 }}
                    >
                        <motion.div 
                            className="w-16 h-16 bg-white rounded-2xl flex items-center justify-center text-3xl mb-8 shadow-sm border border-slate-100"
                            whileHover={{ scale: 1.1, rotate: -3 }}
                            transition={{ type: "spring", stiffness: 400, damping: 10 }}
                        >
                            <i className="fas fa-globe-americas text-transparent bg-clip-text bg-gradient-to-br from-indigo-500 to-purple-600"></i>
                        </motion.div>
                        <h3 className="text-2xl font-bold text-slate-900 mb-4">Our Vision</h3>
                        <p className="text-slate-600 text-lg leading-relaxed">
                            A world where business operations are autonomous, transparent, and intelligent. We envision an ecosystem where compliance is automatic, and strategic growth is the only manual input required.
                        </p>
                    </motion.div>
                </div>
            </div>
        </section>
    );
};