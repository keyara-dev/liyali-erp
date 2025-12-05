import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import * as React from "react";

type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  id?: string;
  label?: string;
  name?: string;
  placeholder?: string;
  onError?: boolean;
  error?: string;
  errorText?: string;
  descriptionText?: string;
  isDisabled?: boolean;
  isInvalid?: boolean;
  onValueChange?: (value: string, meta: any) => void;
  classNames?: {
    wrapper?: string;
    input?: string;
    label?: string;
    errorText?: string;
    descriptionText?: string;
  };
};

const PhoneInput = React.forwardRef<HTMLInputElement, InputProps>(
  (
    {
      id,
      label,
      name,
      placeholder,
      className,
      classNames,
      onError,
      error,
      isInvalid,
      isDisabled,
      descriptionText,
      onValueChange,
      errorText = "",
      ...props
    },
    ref
  ) => {
    return (
      <div
        className={cn(
          "flex w-full max-w-lg flex-col",
          classNames?.wrapper,
          {
            "cursor-not-allowed opacity-50": isDisabled,
          }
        )}
      >
        {label && (
          <label
            className={cn(
              "pl-1 text-sm font-medium text-nowrap text-slate-900/80 mb-0.5",
              {
                "text-red-500": onError || isInvalid,
                "opacity-50": isDisabled || props?.disabled,
              }
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
          id={id}
          name={name}
          placeholder={placeholder}
          type="tel"
          disabled={isDisabled || props?.disabled}
          className={cn(
            "flex items-center h-12 text-lg rounded-lg border border-gray-200 bg-white px-3 text-foreground placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-black focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50",
            {
              "border-red-500 focus:border-red-500/70 focus-visible:ring-red-500/30":
                onError || isInvalid,
            },
            className,
            classNames?.input
          )}
          {...props}
        />

        {((errorText && (isInvalid || onError)) || descriptionText) && (
          <motion.span
            className={cn(
              "ml-1 text-xs text-gray-500",
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
            {errorText ? errorText : descriptionText}
          </motion.span>
        )}
      </div>
    );
  }
);

PhoneInput.displayName = "PhoneInput";

export { PhoneInput };
