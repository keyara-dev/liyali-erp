"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Button } from "./button";

interface DropdownProps {
  trigger: React.ReactNode;
  children: React.ReactNode;
  align?: "left" | "right";
}

export function Dropdown({
  trigger,
  children,
  align = "right",
}: DropdownProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const dropdownRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div className="relative" ref={dropdownRef}>
      <Button
        size="sm"
        variant="ghost"
        onClick={() => setIsOpen(!isOpen)}
        className="p-2 rounded-md hover:bg-gray-100 transition-colors"
      >
        {trigger}
      </Button>

      {isOpen && (
        <div
          className={cn(
            "fixed bg-white rounded-md shadow-xl border border-gray-200 py-1 z-[9999] min-w-[140px] max-w-sm",
            align === "right" ? "right-0" : "left-0"
          )}
          style={{
            top:
              typeof window !== "undefined" && dropdownRef.current
                ? dropdownRef.current.getBoundingClientRect().bottom + 4
                : 0,
            left:
              align === "right"
                ? typeof window !== "undefined" && dropdownRef.current
                  ? dropdownRef.current.getBoundingClientRect().right - 140
                  : 0
                : typeof window !== "undefined" && dropdownRef.current
                  ? dropdownRef.current.getBoundingClientRect().left
                  : 0,
          }}
        >
          {children}
        </div>
      )}
    </div>
  );
}

interface DropdownItemProps {
  onClick: () => void;
  children: React.ReactNode;
  variant?: "default" | "destructive";
}

export function DropdownItem({
  onClick,
  children,
  variant = "default",
}: DropdownItemProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "w-full text-left px-3 py-2 text-sm hover:bg-gray-100 transition-colors",
        variant === "default" && "text-gray-900 hover:text-gray-900",
        variant === "destructive" && "text-red-600 hover:bg-red-50"
      )}
    >
      {children}
    </button>
  );
}
