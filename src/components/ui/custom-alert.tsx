import { cn } from "@/lib/utils";
import React from "react";

type CustomAlertProps = {
  type?: "info" | "warning" | "error" | "success";
  message?: string;
  children?: any;
  className?: string;
  Icon?: React.FC<React.SVGProps<SVGSVGElement>> | null;
};

function CustomAlert({
  type,
  className,
  message,
  Icon,
  children,
}: CustomAlertProps) {
  const getIcon = () => {
    switch (type) {
      case "info":
        return "ℹ️"; // Info icon
      case "warning":
        return "⚠️"; // Warning icon
      case "error":
        return "❌"; // Error icon
      case "success":
        return "✅"; // Success icon
      default:
        return null;
    }
  };

  return (
    <div
      className={cn(
        `p-4 flex items-start gap-4 bg-slate-100 text-black rounded-lg mb-4`,
        {
          "bg-blue-50 text-blue-800 border border-blue-200": type === "info",
          "bg-yellow-50 text-yellow-800 border border-yellow-200":
            type === "warning",
          "bg-red-50 text-red-800 border border-red-200": type === "error",
          "bg-green-50 text-green-800 border border-green-200":
            type === "success",
        },
        className
      )}
    >
      {Icon ? <Icon className="w-6 h-6" /> : getIcon()}
      <div>{message || children}</div>
    </div>
  );
}

export default CustomAlert;
