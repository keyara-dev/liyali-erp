import { TextareaHTMLAttributes, forwardRef } from "react";

import { cn } from "@bgs-tickety/shared";
import { motion } from "framer-motion";

type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label?: string;
  onError?: boolean;
  error?: string;
  errorText?: string;
  descriptionText?: string;
  isInvalid?: boolean;
  showLimit?: boolean;
  classNames?: {
    base?: string;
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
      className,
      onError,
      errorText,
      classNames,
      descriptionText,
      showLimit = false,
      ...props
    },
    ref
  ) => {
    return (
      <div className={cn("w-full", classNames?.wrapper)}>
        {label && (
          <label
            className={cn(
              "pl-1 text-sm font-medium text-nowrap text-foreground/70",
              {
                "text-red-500": onError || props?.isInvalid,
                "opacity-50": props?.disabled,
              }
            )}
            htmlFor={props?.name}
          >
            {label}{" "}
            {props?.required && (
              <span className="font-bold text-red-500"> *</span>
            )}
          </label>
        )}
        <textarea
          ref={ref}
          name={props?.name}
          data-slot="textarea"
          disabled={props?.disabled}
          className={cn(
            "flex min-h-[80px] w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-black ring-offset-white placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-black focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
            {
              "border-red-500 focus:border-red-500/70 focus-visible:ring-red-500/30":
                onError,
              "opacity-50": props?.disabled,
            },
            className,
            classNames?.base
          )}
          {...props}
        />

        {(errorText || descriptionText) && (
          <motion.span
            className={cn(
              "ml-1 text-xs text-muted-foreground flex items-center justify-between gap-2",
              {
                "text-red-600": onError,
              },
              classNames?.descriptionText,
              classNames?.errorText
            )}
            whileInView={{
              scale: [0, 1],
              opacity: [0, 1],
              transition: { duration: 0.3 },
            }}
          >
            {errorText ? errorText : descriptionText}{" "}
            {showLimit && (
              <span>
                {props?.value?.toString()?.length}/{props?.maxLength}
              </span>
            )}
          </motion.span>
        )}
      </div>
    );
  }
);

Textarea.displayName = "Textarea";

export { Textarea };
