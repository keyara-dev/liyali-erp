'use client';

import React from 'react';
import { motion } from 'framer-motion';
import Link from 'next/link';

const LiyaliLogo = () => (
    <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" className="shadow-lg shadow-blue-500/30 rounded-lg">
        <rect width="32" height="32" rx="8" fill="#2563EB"/>
        <path d="M16 26V18M16 18C16 14 13 13 13 13C11 13 10 15 10 18C10 18 10 20 12 20M16 18C16 14 19 13 19 13C21 13 22 15 22 18C22 18 22 20 20 20M16 18V14M16 14C16 10 13 9 13 9C11 9 10 11 10 14M16 14C16 10 19 9 19 9C21 9 22 11 22 14" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        <path d="M16 26H12M16 26H20" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
    </svg>
);

export const Footer = () => {
    const footerSections = [
        {
            title: "Product",
            links: [
                { name: "Features", href: "#features" },
                { name: "Integrations", href: "#" },
                { name: "Pricing", href: "#pricing" },
                { name: "Security", href: "#" }
            ]
        },
        {
            title: "Company",
            links: [
                { name: "About Us", href: "#about" },
                { name: "Careers", href: "#" },
                { name: "Blog", href: "#" },
                { name: "Contact", href: "#" }
            ]
        },
        {
            title: "Resources",
            links: [
                { name: "Documentation", href: "#" },
                { name: "API Reference", href: "#" },
                { name: "Help Center", href: "#" },
                { name: "Community", href: "#" }
            ]
        }
    ];

    const socialLinks = [
        { icon: "fab fa-twitter", href: "#", label: "Twitter" },
        { icon: "fab fa-linkedin-in", href: "#", label: "LinkedIn" },
        { icon: "fab fa-instagram", href: "#", label: "Instagram" }
    ];

    return (
        <footer className="bg-[#020617] text-slate-300 py-16 border-t border-slate-800/50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="grid md:grid-cols-4 gap-12 mb-12">
                    <motion.div 
                        className="col-span-1 md:col-span-1"
                        initial={{ opacity: 0, y: 20 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.6 }}
                    >
                        <motion.div 
                            className="flex items-center gap-2 mb-6"
                            whileHover={{ scale: 1.05 }}
                            transition={{ type: "spring", stiffness: 400, damping: 10 }}
                        >
                            <LiyaliLogo />
                            <span className="text-xl font-bold text-white">Liyali</span>
                        </motion.div>
                        <p className="text-slate-500 text-sm leading-relaxed mb-6">
                            Empowering modern businesses with the tools they need to succeed in a digital-first world.
                        </p>
                        <div className="flex space-x-4">
                            {socialLinks.map((social, index) => (
                                <motion.a 
                                    key={social.label}
                                    href={social.href} 
                                    className="w-9 h-9 rounded-full bg-slate-800 flex items-center justify-center hover:bg-primary-600 hover:text-white transition-all"
                                    whileHover={{ scale: 1.1, y: -2 }}
                                    whileTap={{ scale: 0.95 }}
                                    initial={{ opacity: 0, scale: 0 }}
                                    whileInView={{ opacity: 1, scale: 1 }}
                                    viewport={{ once: true }}
                                    transition={{ duration: 0.3, delay: index * 0.1 }}
                                    aria-label={social.label}
                                >
                                    <i className={social.icon}></i>
                                </motion.a>
                            ))}
                        </div>
                    </motion.div>
                    
                    {footerSections.map((section, sectionIndex) => (
                        <motion.div 
                            key={section.title}
                            initial={{ opacity: 0, y: 20 }}
                            whileInView={{ opacity: 1, y: 0 }}
                            viewport={{ once: true }}
                            transition={{ duration: 0.6, delay: (sectionIndex + 1) * 0.1 }}
                        >
                            <h4 className="text-white font-bold mb-6">{section.title}</h4>
                            <ul className="space-y-3 text-sm text-slate-500">
                                {section.links.map((link, linkIndex) => (
                                    <motion.li 
                                        key={link.name}
                                        initial={{ opacity: 0, x: -10 }}
                                        whileInView={{ opacity: 1, x: 0 }}
                                        viewport={{ once: true }}
                                        transition={{ duration: 0.3, delay: (sectionIndex + 1) * 0.1 + linkIndex * 0.05 }}
                                    >
                                        {link.href.startsWith('#') ? (
                                            <a 
                                                href={link.href} 
                                                className="hover:text-primary-400 transition-colors inline-block"
                                            >
                                                {link.name}
                                            </a>
                                        ) : (
                                            <Link 
                                                href={link.href}
                                                className="hover:text-primary-400 transition-colors inline-block"
                                            >
                                                {link.name}
                                            </Link>
                                        )}
                                    </motion.li>
                                ))}
                            </ul>
                        </motion.div>
                    ))}
                </div>
                
                <motion.div 
                    className="border-t border-slate-900 pt-8 flex flex-col md:flex-row justify-between items-center text-sm text-slate-600"
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    viewport={{ once: true }}
                    transition={{ duration: 0.6, delay: 0.5 }}
                >
                    <p>&copy; {new Date().getFullYear()} Liyali Suite. All rights reserved.</p>
                    <div className="flex space-x-6 mt-4 md:mt-0">
                         <motion.a 
                            href="#" 
                            className="hover:text-white transition-colors"
                            whileHover={{ y: -1 }}
                         >
                            Privacy Policy
                         </motion.a>
                         <motion.a 
                            href="#" 
                            className="hover:text-white transition-colors"
                            whileHover={{ y: -1 }}
                         >
                            Terms of Service
                         </motion.a>
                         <motion.a 
                            href="#" 
                            className="hover:text-white transition-colors"
                            whileHover={{ y: -1 }}
                         >
                            Cookie Policy
                         </motion.a>
                    </div>
                </motion.div>
            </div>
        </footer>
    );
};