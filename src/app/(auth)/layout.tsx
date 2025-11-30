import Logo from "@/components/base/logo";
import { ArrowLeftIcon, ShoppingBagIcon } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import React, { PropsWithChildren } from "react";

function AuthLayout({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen w-screen bg-white">
      {/* Navigation */}
      {/* 
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
      </nav> */}

      <div className="h-screen flex flex-row p-4">
        {/* Left Panel - Auth Forms */}
        <div
          className="bg-primary-700 overflow-clip hidden rounded-[50px] lg:flex w-full items-center justify-center relative"
          // style={{
          //   background:
          //     "linear-gradient(135deg, #111827 0%, #374151 50%, #000000 100%)",
          // }}
        >
          <Image
            src="/images/pattern.svg"
            alt="liyali-pattern"
            width={800}
            height={800}
            className="w-full aspect-square object-cover opacity-20! absolute inset-0 h-full"
          />
          <div className="h-full w-full flex items-center justify-center p-8 relative overflow-clip">
            <div className="w-96 h-96 relative z-10 aspect-square  rounded-3xl overflow-clip justify-center px-8 pl-12 grid place-items-center">
              <Image
                src="/images/logo/logo-icon-plain.svg"
                alt="liyali-logo"
                width={400}
                height={400}
                className="w-full aspect-square object-contain"
              />
            </div>

            {/* Background Elements */}
            <div className="absolute top-20 left-8 w-60 h-60 bg-white/10 rounded-full blur-xl"></div>
            <div className="absolute bottom-20 right-20 w-48 h-48 bg-white/5 rounded-full blur-2xl"></div>
            <div className="absolute top-1/2 left-24 w-48 h-48 bg-white/10 rounded-full blur-lg"></div>
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
