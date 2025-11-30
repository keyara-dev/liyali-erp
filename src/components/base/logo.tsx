"use client";
import { useTheme } from "next-themes";
import Link from "next/link";
import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";
import Image from "next/image";
import { Skeleton } from "../ui/skeleton";

type LogoProps = {
  href?: string;
  src?: string;
  alt?: string;
  name?: string;
  width?: number;
  height?: number;
  className?: string;
  isIcon?: boolean;
  isWhite?: boolean;
  isDark?: boolean;
  isFull?: boolean;
};

function Logo({
  href = "/",
  src,
  alt,
  isWhite = false,
  isDark = false,
  isFull = true,
  className = "",
  isIcon = false,
}: LogoProps) {
  const { theme } = useTheme();
  const [logoUrl, setLogoUrl] = useState("/images/logo/logo-full.svg");

  useEffect(() => {
    let logoType: string;

    if (isIcon) {
      logoType =
        theme === "light" || theme === "dark"
          ? `/images/logo/logo-${theme}.svg`
          : `/images/logo/logo-icon.svg`;
    } else if (isWhite) {
      logoType = "/images/logo/logo-light.svg";
    } else if (isDark) {
      logoType = "/images/logo/logo-dark.svg";
    } else if (isFull) {
      logoType =
        theme === "light" || theme === "dark"
          ? `/images/logo/logo-full-${theme}.svg`
          : "/images/logo/logo-full-dark.svg";
    } else {
      logoType =
        theme === "light" || theme === "dark"
          ? `/images/logo/logo-${theme}.svg`
          : "/images/logo/logo-dark.svg";
    }

    setLogoUrl(logoType);
  }, [theme, isIcon, isWhite, isDark, isFull]);

  // LOADING STATE
  if (!logoUrl) {
    return <Skeleton className="flex-1 h-9" />;
  }

  // RENDERER
  if (isIcon) {
    return (
      <Link href={href}>
        <div
          className={cn(
            `aspect-square flex justify-center w-full max-h-12 items-center min-w-fit`,
            className,
            {
              "max-w-12 mx-auto max-h-12 min-h-12 ": isIcon,
            }
          )}
        >
          <Image
            alt={alt || "logo"}
            className="object-contain"
            height={50}
            src={logoUrl}
            width={50}
          />
        </div>
      </Link>
    );
  } else {
    return (
      <Link href={href}>
        <div className={cn(`w-full min-w-fit items-center`, className)}>
          <Image
            alt={alt || "logo"}
            className="object-contain transition-all my-auto min-h-8 duration-300 ease-in-out"
            height={60}
            src={src || logoUrl}
            width={160}
          />
        </div>
      </Link>
    );
  }
}

export default Logo;
