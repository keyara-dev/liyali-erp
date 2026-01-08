import * as React from "react";

import { cn } from "@/lib/utils";

export interface SelectProps
  extends React.SelectHTMLAttributes<HTMLSelectElement> {}

const Select = React.forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, ...props }, ref) => {
    return (
      <select
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",    // Base styles
            "w-full px-4 py-2 text-base bg-white border border-slate-200 rounded-lg transition-all duration-200 outline-none",
            // Placeholder styles
            "placeholder:text-slate-400",
            // Focus styles with primary color
            "focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:shadow-lg focus:shadow-primary-500/10",
            // Hover styles
            "hover:border-slate-300",
            // Error styles
            // {
            //   "border-red-500 focus:border-red-500 focus:ring-red-500/20 focus:shadow-red-500/10":
            //     onError || isInvalid,
            // },
            // Disabled styles
            "disabled:bg-slate-50 disabled:text-slate-500 disabled:cursor-not-allowed disabled:opacity-60",
            // Text styles
            "text-slate-900 selection:bg-primary-100 selection:text-primary-900",
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
