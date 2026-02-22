import * as React from "react";

import { cn } from "@/lib/utils";

export interface SelectProps
  extends React.SelectHTMLAttributes<HTMLSelectElement> {}

const Select = React.forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, ...props }, ref) => {
    return (
      <select
        className={cn(
          // Base styles
            "w-full px-4 py-2 text-base bg-foreground/5 border border-border rounded-lg transition-all duration-200 outline-none",
            // Placeholder styles
            "placeholder:text-slate-400 dark:placeholder:text-slate-500",
            // Focus styles with primary color
            "focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:shadow-lg focus:shadow-primary-500/10",
            // Hover styles
            "hover:border-slate-300/50",
            // Disabled styles
            "disabled:bg-slate-50/50 disabled:text-slate-500 disabled:cursor-not-allowed disabled:opacity-60",
            // Text styles
            "text-slate-900 dark:text-slate-100 selection:bg-primary-100 selection:text-primary-900",
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
Select.displayName = "Select";

export { Select };
