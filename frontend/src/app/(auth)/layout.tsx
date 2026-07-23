import Image from "next/image";
import Link from "next/link";
import { PropsWithChildren } from "react";

export const dynamic = "force-dynamic";

function AuthLayout({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen w-screen ">
      {/* Back to Home Button - Top Right */}
      <div className="absolute top-4 right-4 z-50">
        <Link href="/">
          <button className="flex items-center gap-2 px-4 py-2 bg-white/90 hover:bg-white backdrop-blur-sm border border-slate-200 rounded-full text-slate-700 hover:text-slate-900 transition-all duration-300 shadow-lg hover:shadow-xl hover:scale-105 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2">
            <i className="fas fa-arrow-left text-sm"></i>
            <span className="text-sm font-medium">Back to Home</span>
          </button>
        </Link>
      </div>

      <div className="h-screen flex flex-row p-4">
        {/* Left Panel - Auth Forms */}
        <div className="bg-primary-700 overflow-clip hidden rounded-[50px] lg:flex w-full items-center justify-center relative">
          <Image
            src="/images/pattern.svg"
            alt="liyali-pattern"
            width={800}
            height={800}
            className="w-full bg- aspect-square object-cover opacity-20! absolute inset-0 h-full"
          />
          <div className="h-full w-full flex items-center justify-center relative overflow-clip">
            <div className="w-96 h-96 relative z-10 aspect-square  rounded-3xl overflow-clip justify-center px-8 pl-12 grid place-items-center">
              <Image
                src="/images/logo/logo-icon-plain.svg"
                alt="liyali-logo"
                width={600}
                height={600}
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

          <div className="mt-auto pt-8 border-t border-foreground/10 text-center text-slate-500 text-sm">
            <p>
              &copy; {new Date().getFullYear()} Liyali Suite. All rights
              reserved.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default AuthLayout;
