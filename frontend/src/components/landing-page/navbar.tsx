"use client";

import React, { useState, useEffect } from "react";
import { motion } from "framer-motion";
import Link from "next/link";
import Logo from "../base/logo";
import { cn } from "@/lib/utils";

const LiyaliLogo = () => (
  <svg
    width="32"
    height="32"
    viewBox="0 0 32 32"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    className="shadow-lg shadow-blue-500/30 rounded-lg"
  >
    <rect width="32" height="32" rx="8" fill="#2563EB" />
    <path
      d="M16 26V18M16 18C16 14 13 13 13 13C11 13 10 15 10 18C10 18 10 20 12 20M16 18C16 14 19 13 19 13C21 13 22 15 22 18C22 18 22 20 20 20M16 18V14M16 14C16 10 13 9 13 9C11 9 10 11 10 14M16 14C16 10 19 9 19 9C21 9 22 11 22 14"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M16 26H12M16 26H20"
      stroke="white"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

export const Navbar = ({ isAuthenticated }: { isAuthenticated: boolean }) => {
  const [scrolled, setScrolled] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [activeSection, setActiveSection] = useState("");

  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);

      // Scroll Spy Logic
      const sections = ["features", "how-it-works", "pricing", "about"];
      const scrollY = window.scrollY;
      // Activate section when it's in the top third of the screen
      const offset = window.innerHeight * 0.3;

      let current = "";
      for (const section of sections) {
        const element = document.getElementById(section);
        if (element) {
          const top = element.offsetTop - 120; // Account for navbar height + buffer
          const bottom = top + element.offsetHeight;
          if (scrollY + offset >= top && scrollY + offset < bottom) {
            current = section;
            break;
          }
        }
      }
      setActiveSection(current);
    };

    window.addEventListener("scroll", handleScroll);
    handleScroll(); // Check on mount
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const getLinkClass = (sectionId: string) => {
    const baseClass =
      "text-sm font-medium transition-all duration-300 relative px-1 py-1 cursor-pointer";
    // Active: White text, visible blue dot
    const activeClass =
      "text-primary hover:text-primary/80 font-bold after:content-[''] after:absolute after:-bottom-1 after:left-1/2 after:-translate-x-1/2 after:w-1.5 after:h-1.5 after:bg-blue-500 after:rounded-full after:opacity-100 after:transition-all after:duration-300";
    // Inactive: Primary text, invisible dot (grows on hover if you want, or just fades in)
    const inactiveClass =
      "text-foreground/80 hover:text-primary after:content-[''] after:absolute after:-bottom-1 after:left-1/2 after:-translate-x-1/2 after:w-0 after:h-0 after:bg-blue-500 after:rounded-full after:opacity-0 hover:after:opacity-50 hover:after:w-1 hover:after:h-1";

    return `${baseClass} ${activeSection === sectionId ? activeClass : inactiveClass}`;
  };

  return (
    <>
      <motion.nav
        initial={{ y: -100, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
        className={cn(
          `fixed top-2 left-2 right-2 sm:top-4 sm:left-4 sm:right-4 md:left-1/2 md:-translate-x-1/2 w-full container max-w-7xl z-50 transition-all duration-300 py-3 sm:py-4`,
          {
            "bg-card/80 rounded-2xl sm:rounded-full backdrop-blur-md py-2 sm:py-3 border border-border/50 shadow-2xl md:w-auto md:min-w-[600px] md:max-w-5xl transition-all duration-400 ":
              scrolled,
          },
        )}
      >
        <div className="px-3 sm:px-6 md:px-8">
          <div className="flex justify-between items-center gap-4 sm:gap-8">
            {/* Logo */}
            <motion.div
              className="flex items-center gap-2 shrink-0"
              whileHover={{ scale: 1.05 }}
              transition={{ type: "spring", stiffness: 400, damping: 10 }}
            >
              <Logo
                isFull
                href="/"
                width={scrolled ? 60 : 100}
                height={24}
                classNames={{
                  image: "max-h-6 sm:max-h-8 md:max-h-9 2xl:max-h-10",
                }}
              />
            </motion.div>

            {/* Desktop Links - Centered */}
            <div className="hidden md:flex items-center justify-center space-x-8 flex-1">
              <Link href="#features" className={getLinkClass("features")}>
                Product
              </Link>
              <Link
                href="#how-it-works"
                className={getLinkClass("how-it-works")}
              >
                How it Works
              </Link>
              <Link href="#pricing" className={getLinkClass("pricing")}>
                Pricing
              </Link>
              <Link href="#about" className={getLinkClass("about")}>
                About
              </Link>
            </div>

            {/* CTA - Right */}
            <div className="hidden md:flex items-center flex-shrink-0">
              {
                <Link href={isAuthenticated ? "/home" : "/login"}>
                  <motion.button
                    className="bg-primary-600 cursor-pointer hover:bg-primary-500 text-white px-6 py-2 rounded-full text-sm font-bold transition-all shadow-[0_0_20px_rgba(37,99,235,0.3)] hover:shadow-[0_0_25px_rgba(37,99,235,0.5)] border border-primary-500/50"
                    whileHover={{ y: -2 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    {isAuthenticated ? "Go to Dashboard" : "Login"}
                  </motion.button>
                </Link>
              }
            </div>

            {/* Mobile Toggle */}
            <div className="md:hidden">
              <button
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                className="text-slate-600 hover:text-slate-900 transition-colors p-2 rounded-lg hover:bg-slate-100"
                aria-label="Toggle mobile menu"
              >
                <i
                  className={`fas ${mobileMenuOpen ? "fa-times" : "fa-bars"} text-lg`}
                ></i>
              </button>
            </div>
          </div>
        </div>
      </motion.nav>

      {/* Mobile Menu Overlay */}
      {mobileMenuOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-40 bg-white/95 backdrop-blur-xl pt-20 px-4 md:hidden"
        >
          <motion.div
            initial={{ y: 50, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            transition={{ delay: 0.1 }}
            className="flex flex-col space-y-8 text-center max-w-sm mx-auto"
          >
            <a
              href="#features"
              onClick={() => setMobileMenuOpen(false)}
              className="text-xl text-slate-700 font-medium py-3 border-b border-slate-200"
            >
              Product
            </a>
            <a
              href="#how-it-works"
              onClick={() => setMobileMenuOpen(false)}
              className="text-xl text-slate-700 font-medium py-3 border-b border-slate-200"
            >
              How it Works
            </a>
            <a
              href="#pricing"
              onClick={() => setMobileMenuOpen(false)}
              className="text-xl text-slate-700 font-medium py-3 border-b border-slate-200"
            >
              Pricing
            </a>
            <a
              href="#about"
              onClick={() => setMobileMenuOpen(false)}
              className="text-xl text-slate-700 font-medium py-3 border-b border-slate-200"
            >
              About
            </a>
            <div className="pt-4">
              <Link href={isAuthenticated ? "/home" : "/login"}>
                <button className="bg-primary-600 hover:bg-primary-700 text-white py-4 px-8 rounded-xl font-bold text-lg shadow-lg shadow-primary-500/30 w-full">
                  {isAuthenticated ? "Go to Dashboard" : "Login"}
                </button>
              </Link>
            </div>
          </motion.div>
        </motion.div>
      )}
    </>
  );
};
