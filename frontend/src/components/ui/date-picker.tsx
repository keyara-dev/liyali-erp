"use client";

import * as React from "react";
import { format } from "date-fns";
import { ChevronDownIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";

export function DatePicker({
  label,
  name,
  isDisabled,
  isInvalid,
  onError,
  errorText,
  descriptionText,
  classNames,
  value,
  onValueChange,
  minDate,
  maxDate,
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  onError?: boolean;
  error?: string;
  errorText?: string;
  descriptionText?: string;
  isDisabled?: boolean;
  isInvalid?: boolean;
  value?: Date;
  onValueChange?: (value?: Date) => void;
  minDate?: Date;
  maxDate?: Date;
  classNames?: {
    wrapper?: string;
    input?: string;
    label?: string;
    errorText?: string;
    descriptionText?: string;
  };
}) {
  const [open, setOpen] = React.useState(false);

  // Create disabled matcher for dates
  const disabledDates = React.useMemo(() => {
    if (!minDate && !maxDate) return undefined;

    return (date: Date) => {
      if (minDate && date < minDate) return true;
      if (maxDate && date > maxDate) return true;
      return false;
    };
  }, [minDate, maxDate]);

  return (
    <div className={cn("flex w-full flex-col", classNames?.wrapper)}>
      {label && (
        <label
          className={cn("mb-0.5 pl-1 text-sm font-medium", classNames?.label, {
            "text-red-500": isInvalid,
            "opacity-50": props?.disabled
          })}
          htmlFor={name}>
          {label} {props?.required && <span className="font-bold text-red-500"> *</span>}
        </label>
      )}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            id="date"
            className={cn(
              "w-full justify-between text-left font-normal",
              !value && "text-muted-foreground",
              classNames?.input
            )}
            disabled={isDisabled || props?.disabled}>
            {value && value instanceof Date && !isNaN(value.getTime()) ? (
              format(value, "PPP")
            ) : (
              <span>Pick a date</span>
            )}
            <ChevronDownIcon />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-0" align="start">
          <Calendar
            mode="single"
            selected={value}
            disabled={disabledDates || isDisabled || props?.disabled}
            captionLayout="dropdown"
            startMonth={new Date(1900, 0)}
            endMonth={new Date(2099, 11)}
            onSelect={(date) => {
              onValueChange && onValueChange(date);
              setOpen(false);
            }}
          />
        </PopoverContent>
      </Popover>
      {((errorText && (isInvalid || onError)) || descriptionText) && (
        <motion.span
          className={cn(
            "ml-1 text-xs text-gray-500",
            {
              "text-red-600": onError || isInvalid
            },
            classNames?.descriptionText,
            classNames?.errorText
          )}
          whileInView={{
            scale: [0, 1],
            opacity: [0, 1],
            transition: { duration: 0.3 }
          }}>
          {errorText ? errorText : descriptionText}
        </motion.span>
      )}
    </div>
  );
}
