import { TextareaHTMLAttributes, forwardRef } from "react";

import { cn } from "@/lib/utils";
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
      isInvalid,
      showLimit = false,
      ...props
    },
    ref
  ) => {
    return (
      <div className={cn("w-full", classNames?.wrapper)}>
        {label && (
          <label
            className={cn("mb-0.5 pl-1 text-sm font-medium text-nowrap", {
              "text-red-500": onError || isInvalid,
              "opacity-50": props?.disabled,
            })}
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
            "dark:bg-input/20 border-input text-foreground ring-offset-foreground placeholder:text-foreground/50 flex min-h-[60px] w-full rounded-lg border px-3 py-2 text-sm focus-visible:ring-1 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
            {
              "border-red-500 focus:border-red-500/70 focus-visible:ring-red-500/30":
                onError || isInvalid,
              "opacity-50": props?.disabled,
            },
            className,
            classNames?.base
          )}
          {...props}
        />

        {((errorText && isInvalid) || descriptionText) && (
          <motion.span
            className={cn(
              "text-muted-foreground ml-1 flex items-center justify-between gap-2 text-xs",
              {
                "text-red-600": onError || isInvalid,
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
