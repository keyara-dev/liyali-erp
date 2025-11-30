import Logo from "@/components/base/logo";
import { ArrowLeftIcon, ShoppingBagIcon } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import React, { PropsWithChildren } from "react";

function AuthLayout({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen w-screen bg-white">
      {/* Navigation */}

      <nav className="px-8 md:px-12 py-1 w-screen  flex justify-between fixed top-0 left-0 right-0 items-center">
        <Link href="/" className="text-black font-bold text-2xl">
          <div className="flex items-center">
            <Image
              src="/logo/logo.png"
              alt="xclsv"
              width={120}
              height={40}
              className="h-18 w-auto"
            />
          </div>
        </Link>
        <Link
          href="/"
          className="px-4 bg-white/60 text-foreground/80 sm:px-6 py-2 rounded-full font-medium text-lg hover:opacity-80 transition-all backdrop-blur-lg ease-in-out duration-300"
        >
          <ArrowLeftIcon className="inline w-6 h-6 mr-2" />
          <span className="hidden sm:inline">Go back </span>Home
        </Link>
      </nav>

      <div className="h-screen  bg-background flex flex-row">
        {/* Left Panel - Auth Forms */}
        <div
          className="hidden lg:flex w-full items-center justify-center"
          style={{
            background:
              "linear-gradient(135deg, #111827 0%, #374151 50%, #000000 100%)",
          }}
        >
          <div className="h-full flex items-center justify-center p-8 relative overflow-hidden">
            <div className="relative z-10">
              <div className="w-80 h-80 bg-gradient-to-br from-gray-800 to-black rounded-3xl shadow-2xl flex items-center justify-center transform rotate-12 hover:rotate-6 transition-transform duration-200">
                <div className="w-64 h-64 bg-gradient-to-br from-gray-700 to-gray-900 rounded-2xl shadow-inner flex items-center justify-center">
                  <div className="w-32 h-32 bg-white rounded-full shadow-lg flex items-center justify-center">
                    <Logo isIcon className="w-16 h-16 text-black" />
                  </div>
                </div>
              </div>
            </div>

            {/* Background Elements */}
            <div className="absolute top-20 left-20 w-32 h-32 bg-white/10 rounded-full blur-xl"></div>
            <div className="absolute bottom-20 right-20 w-48 h-48 bg-white/5 rounded-full blur-2xl"></div>
            <div className="absolute top-1/2 left-10 w-24 h-24 bg-white/20 rounded-full blur-lg"></div>
          </div>
        </div>

        {/* Right Panel - Illustration */}
        <div className="w-full p-8 pt-32 overflow-y-auto flex flex-col">
          <div className="flex flex-col justify-center items-center">
            {children}
          </div>
        </div>
      </div>
    </div>
  );
}

export default AuthLayout;
