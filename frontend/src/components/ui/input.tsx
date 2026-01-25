import { cn } from "@/lib/utils";
import * as React from "react";
import { motion } from "framer-motion";

type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  name?: string;
  onError?: boolean;
  error?: string;
  errorText?: string;
  descriptionText?: string;
  isDisabled?: boolean;
  isInvalid?: boolean;
  classNames?: {
    wrapper?: string;
    input?: string;
    label?: string;
    errorText?: string;
    descriptionText?: string;
  };
};

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      type,
      label,
      name,
      classNames,
      onError,
      error,
      maxLength,
      max,
      isInvalid,
      min,
      isDisabled,
      descriptionText,
      errorText = "",
      ...props
    },
    ref,
  ) => {
    return (
      <div
        className={cn("flex w-full flex-col", classNames?.wrapper, {
          "cursor-not-allowed opacity-50": isDisabled,
        })}
      >
        {label && (
          <label
            className={cn(
              "mb-0.5 text-sm font-medium text-slate-700",
              {
                "text-red-500": onError || isInvalid,
                "opacity-50": isDisabled || props?.disabled,
              },
              classNames?.label,
            )}
            htmlFor={name}
          >
            {label}{" "}
            {props?.required && (
              <span className="font-bold text-red-500"> *</span>
            )}
          </label>
        )}
        <input
          ref={ref}
          className={cn(
            // Base styles
            "w-full px-4 py-2 text-base bg-white border border-slate-200 rounded-lg transition-all duration-200 outline-none",
            // Placeholder styles
            "placeholder:text-slate-400",
            // Focus styles with primary color
            "focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:shadow-lg focus:shadow-primary-500/10",
            // Hover styles
            "hover:border-slate-300",
            // Error styles
            {
              "border-red-500 focus:border-red-500 focus:ring-red-500/20 focus:shadow-red-500/10":
                onError || isInvalid,
            },
            // Disabled styles
            "disabled:bg-slate-50 disabled:text-slate-500 disabled:cursor-not-allowed disabled:opacity-60",
            // Text styles
            "text-slate-900 selection:bg-primary-100 selection:text-primary-900",
            className,
            classNames?.input,
          )}
          disabled={isDisabled || props?.disabled}
          id={name}
          maxLength={maxLength}
          max={max}
          min={min}
          name={name}
          type={type}
          {...props}
        />

        {((errorText && (isInvalid || onError)) || descriptionText) && (
          <motion.span
            className={cn(
              "ml-1 text-xs text-slate-500",
              {
                "text-red-600": onError || isInvalid,
              },
              classNames?.descriptionText,
              classNames?.errorText,
            )}
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.2 }}
          >
            {errorText ? errorText : descriptionText}
          </motion.span>
        )}
      </div>
    );
  },
);

Input.displayName = "Input";

export { Input };
