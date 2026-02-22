import { TextareaHTMLAttributes, forwardRef } from "react";

import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label?: string;
  name?: string;
  onError?: boolean;
  error?: string;
  errorText?: string;
  descriptionText?: string;
  isDisabled?: boolean;
  isInvalid?: boolean;
  showLimit?: boolean;
  classNames?: {
    wrapper?: string;
    input?: string;
    label?: string;
    errorText?: string;
    descriptionText?: string;
  };
};

const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  (
    {
      label,
      name,
      className,
      onError,
      errorText,
      classNames,
      descriptionText,
      isDisabled,
      isInvalid,
      showLimit = false,
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
              "mb-1 text-sm font-medium text-slate-700 dark:text-slate-300",
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
        <textarea
          ref={ref}
          className={cn(
            // Base styles matching input.tsx
            "w-full px-4 py-2 text-base bg-foreground/5 border border-border rounded-lg transition-all duration-200 outline-none",
            // Placeholder styles
            "placeholder:text-slate-400 dark:placeholder:text-slate-500",
            // Focus styles with primary color
            "focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:shadow-lg focus:shadow-primary-500/10",
            // Hover styles
            "hover:border-slate-300/50",
            // Error styles
            {
              "border-red-500 focus:border-red-500 focus:ring-red-500/20 focus:shadow-red-500/10":
                onError || isInvalid,
            },
            // Disabled styles
            "disabled:bg-slate-50/50 disabled:text-slate-500 disabled:cursor-not-allowed disabled:opacity-60",
            // Text styles
            "text-slate-900 dark:text-slate-100 selection:bg-primary-100 selection:text-primary-900",
            // Textarea specific styles
            "min-h-[80px] resize-vertical",
            className,
            classNames?.input,
          )}
          disabled={isDisabled || props?.disabled}
          id={name}
          name={name}
          {...props}
        />

        {((errorText && (isInvalid || onError)) || descriptionText) && (
          <motion.span
            className={cn(
              "ml-1 flex items-center justify-between gap-2 text-xs text-slate-500 dark:text-slate-400",
              {
                "text-red-600 dark:text-red-400": onError || isInvalid,
              },
              classNames?.descriptionText,
              classNames?.errorText,
            )}
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            transition={{ duration: 0.2 }}
          >
            <span>{errorText ? errorText : descriptionText}</span>
            {showLimit && (
              <span className="text-slate-400 dark:text-slate-500">
                {props?.value?.toString()?.length || 0}/{props?.maxLength}
              </span>
            )}
          </motion.span>
        )}
      </div>
    );
  },
);

Textarea.displayName = "Textarea";

export { Textarea };
