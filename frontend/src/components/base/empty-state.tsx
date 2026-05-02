"use client";

import { cn } from "@/lib/utils";

import { motion } from "framer-motion";
import { CircleHelpIcon } from "lucide-react";

export default function EmptyState({
  Icon = CircleHelpIcon,
  title,
  description,
  action,
  className,
  classNames,
}: {
  Icon?: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
  /** Optional action node, e.g. a Button or Link. Rendered below description. */
  action?: React.ReactNode;
  className?: string;
  classNames?: {
    icon?: string;
    container?: string;
    title?: string;
    description?: string;
    action?: string;
  };
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -12 }}
      transition={{ duration: 0.2, ease: [0.2, 0.8, 0.2, 1] }}
      className={cn(
        "flex w-full flex-col items-center justify-center gap-2 max-w-2xl py-10",
        className,
        classNames?.container
      )}
    >
      {Icon && (
        <Icon
          className={cn("w-12 h-12 text-muted-foreground/60", classNames?.icon)}
          aria-hidden="true"
        />
      )}
      <h4
        className={cn(
          "text-center text-base leading-6 text-foreground font-semibold",
          classNames?.title
        )}
      >
        {title}
      </h4>
      <p
        className={cn(
          "text-center text-xs sm:text-sm text-muted-foreground max-w-md",
          classNames?.description
        )}
      >
        {description}
      </p>
      {action && (
        <div className={cn("mt-2", classNames?.action)}>{action}</div>
      )}
    </motion.div>
  );
}
