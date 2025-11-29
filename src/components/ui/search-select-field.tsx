import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@bgs-tickety/shared";
import { motion } from "framer-motion";
import { Check, ChevronsUpDown } from "lucide-react";
import * as React from "react";

type SelectInputProps = React.InputHTMLAttributes<HTMLSelectElement> & {
  label?: string;
  onError?: boolean;
  error?: string;
  errorText?: string;
  descriptionText?: string;
  isDisabled?: boolean;
  isInvalid?: boolean;
  onModal?: boolean;
  value?: string;
  className?: string;
  listItemName?: string;
  options: { id: string; name: string; [x: string]: any }[];
  onValueChange?: (value: string) => void;
  classNames?: {
    wrapper?: string;
    input?: string;
    label?: string;
    errorText?: string;
    descriptionText?: string;
  };
};

const SearchSelectField = React.forwardRef<HTMLSelectElement, SelectInputProps>(
  (
    {
      className,
      type,
      label,
      name,
      value,
      classNames,
      onError,
      error,
      defaultValue = "",
      placeholder,
      onValueChange,
      listItemName,
      isInvalid,
      options,
      isDisabled,
      descriptionText,
      errorText = "",
      onModal = false,
      ...props
    },
    ref
  ) => {
    const [open, setOpen] = React.useState(false);
    const [selected, setSelected] = React.useState("");

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

        <Popover modal={onModal} open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={open}
              className={cn(
                "justify-between",
                {
                  "cursor-not-allowed opacity-50": isDisabled,
                  "border-red-500": onError || isInvalid,
                  "text-foreground/60": !selected,
                },
                classNames?.input
              )}
            >
              {selected
                ? options.find((item) => item?.id === selected)?.name
                : placeholder || "Select an item..."}
              <ChevronsUpDown className="opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[var(--radix-popover-trigger-width)] flex p-0">
            <Command>
              <CommandInput
                placeholder="Type here to search..."
                className="h-9"
              />
              <CommandList>
                <CommandEmpty>No items found.</CommandEmpty>
                <CommandGroup>
                  {options.map((item) => {
                    const itemValue = item.id;
                    const itemLabel = item?.[String(listItemName)] || item.name;

                    return (
                      <CommandItem
                        key={itemValue}
                        value={itemValue}
                        onSelect={(currentValue) => {
                          const selectedItem =
                            currentValue === value ? "" : currentValue;

                          setSelected(selectedItem);
                          onValueChange?.(selectedItem);
                          setOpen(false);
                        }}
                      >
                        {itemLabel}
                        <Check
                          className={cn(
                            "ml-auto",
                            selected === itemValue ? "opacity-100" : "opacity-0"
                          )}
                        />
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

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

SearchSelectField.displayName = "SearchSelectField";

export { SearchSelectField };
