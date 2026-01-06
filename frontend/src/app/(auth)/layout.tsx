import Image from "next/image";
import { PropsWithChildren } from "react";

function AuthLayout({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen w-screen ">
    

      <div className="h-screen flex flex-row p-4">
        {/* Left Panel - Auth Forms */}
        <div
          className="bg-primary-700 overflow-clip hidden rounded-[50px] lg:flex w-full items-center justify-center relative"
        
        >
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
        </div>
      </div>
    </div>
  );
}

export default AuthLayout;
