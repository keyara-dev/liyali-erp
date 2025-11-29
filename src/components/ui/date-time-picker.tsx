"use client";

import * as React from "react";
import { format } from "date-fns";
import { Calendar as CalendarIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface DateTimePickerProps {
  value?: Date;
  onValueChange?: (date?: Date) => void;
  className?: string;
  label?: string;
  required?: boolean;
}

export function DateTimePicker({
  value,
  onValueChange,
  className,
  label,
  required
}: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false);

  const handleDateSelect = (selectedDate?: Date) => {
    if (!selectedDate || !onValueChange) {
      onValueChange?.(undefined);
      return;
    }

    const currentVal = value || new Date();
    const newDate = new Date(
      selectedDate.getFullYear(),
      selectedDate.getMonth(),
      selectedDate.getDate(),
      currentVal.getHours(),
      currentVal.getMinutes(),
      currentVal.getSeconds()
    );

    onValueChange(newDate);
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const timeValue = e.target.value;
    if (!timeValue || !onValueChange) return;

    const [hours, minutes] = timeValue.split(":").map(Number);
    const currentVal = value || new Date();
    const newDate = new Date(
      currentVal.getFullYear(),
      currentVal.getMonth(),
      currentVal.getDate(),
      hours,
      minutes,
      0 // seconds
    );

    onValueChange(newDate);
  };

  return (
    <div className={cn("grid gap-2", className)}>
      {label && (
        <Label>
          {label} {required && <span className="font-bold text-red-500"> *</span>}
        </Label>
      )}
      <div className="flex items-center gap-2">
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button
              variant={"outline"}
              className={cn(
                "w-[240px] justify-start text-left font-normal",
                !value && "text-muted-foreground"
              )}>
              <CalendarIcon className="mr-2 h-4 w-4" />
              {value ? format(value, "PPP") : <span>Pick a date</span>}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar
              mode="single"
              selected={value}
              onSelect={(date) => {
                handleDateSelect(date);
                setOpen(false);
              }}
              initialFocus
            />
          </PopoverContent>
        </Popover>
        <Input
          type="time"
          step="1"
          value={value ? format(value, "HH:mm") : ""}
          onChange={handleTimeChange}
          className="w-auto"
        />
      </div>
    </div>
  );
}
