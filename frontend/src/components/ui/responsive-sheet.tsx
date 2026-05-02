"use client";
import * as React from "react";
import { Drawer as Vaul } from "vaul";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

function useIsMobile() {
  const [is, setIs] = React.useState(false);
  React.useEffect(() => {
    const mq = window.matchMedia("(max-width: 767px)");
    const update = () => setIs(mq.matches);
    update();
    mq.addEventListener("change", update);
    return () => mq.removeEventListener("change", update);
  }, []);
  return is;
}

export interface ResponsiveSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: React.ReactNode;
  description?: React.ReactNode;
  children: React.ReactNode;
  footer?: React.ReactNode;
  /** Tailwind max-w on desktop. */
  desktopMaxWidth?: string;
  className?: string;
}

export function ResponsiveSheet({
  open,
  onOpenChange,
  title,
  description,
  children,
  footer,
  desktopMaxWidth = "sm:max-w-lg",
  className,
}: ResponsiveSheetProps) {
  const isMobile = useIsMobile();

  if (isMobile) {
    return (
      <Vaul.Root open={open} onOpenChange={onOpenChange}>
        <Vaul.Portal>
          <Vaul.Overlay className="fixed inset-0 bg-black/40 z-50" />
          <Vaul.Content
            className={cn(
              "fixed bottom-0 left-0 right-0 z-50 mt-24 flex max-h-[90svh] flex-col rounded-t-xl bg-background border-t",
              className
            )}
          >
            <div className="mx-auto mt-2 h-1.5 w-12 shrink-0 rounded-full bg-muted" />
            {(title || description) && (
              <div className="px-4 pt-3 pb-2 space-y-1">
                {title && (
                  <Vaul.Title className="text-base font-semibold">
                    {title}
                  </Vaul.Title>
                )}
                {description && (
                  <Vaul.Description asChild>
                    <div className="text-sm text-muted-foreground">
                      {description}
                    </div>
                  </Vaul.Description>
                )}
              </div>
            )}
            <div className="flex-1 overflow-y-auto px-4 pb-4">{children}</div>
            {footer && (
              <div className="border-t p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
                {footer}
              </div>
            )}
          </Vaul.Content>
        </Vaul.Portal>
      </Vaul.Root>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className={cn(desktopMaxWidth, "overflow-y-auto max-h-[90svh] p-0", className)}
      >
        {(title || description) && (
          <DialogHeader className="px-6 pt-5 pb-2">
            {title && <DialogTitle>{title}</DialogTitle>}
            {description && (
              <DialogDescription asChild>
                <div className="text-sm text-muted-foreground">
                  {description}
                </div>
              </DialogDescription>
            )}
          </DialogHeader>
        )}
        <div className="px-6 pb-4">{children}</div>
        {footer && <DialogFooter className="px-6 py-4 border-t">{footer}</DialogFooter>}
      </DialogContent>
    </Dialog>
  );
}
