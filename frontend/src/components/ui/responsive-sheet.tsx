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
  /** Tailwind max-w on desktop. Ignored when `fullScreen` is true. */
  desktopMaxWidth?: string;
  /**
   * Fill the viewport. Desktop: 95vw × 95vh. Mobile: inset-0 (full-screen
   * drawer with top sheet edge instead of a partial pull-up). Use for
   * PDF/image viewers and other read-only content that needs maximum screen
   * area. Default: false.
   */
  fullScreen?: boolean;
  /**
   * When false, clicking the backdrop / pressing Escape / dragging the sheet
   * down does NOT dismiss. Use for forms with unsaved state.
   * Default: true.
   */
  dismissibleOnOutsideClick?: boolean;
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
  fullScreen = false,
  dismissibleOnOutsideClick = true,
  className,
}: ResponsiveSheetProps) {
  const isMobile = useIsMobile();

  if (isMobile) {
    return (
      <Vaul.Root
        open={open}
        onOpenChange={onOpenChange}
        dismissible={dismissibleOnOutsideClick}
      >
        <Vaul.Portal>
          <Vaul.Overlay className="fixed inset-0 bg-black/40 z-50" />
          <Vaul.Content
            className={cn(
              "fixed left-0 right-0 z-50 flex flex-col bg-background border-t",
              fullScreen
                ? "inset-0 rounded-none"
                : "bottom-0 mt-24 max-h-[90svh] rounded-t-xl",
              className
            )}
          >
            {!fullScreen && (
              <div className="mx-auto mt-2 h-1.5 w-12 shrink-0 rounded-full bg-muted" />
            )}
            {(title || description) && (
              <div
                className={cn(
                  "px-4 pb-2 space-y-1 shrink-0",
                  fullScreen ? "pt-4" : "pt-3"
                )}
              >
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
            <div className="flex-1 overflow-y-auto px-4 pb-4 min-h-0">
              {children}
            </div>
            {footer && (
              <div className="shrink-0 border-t bg-background p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
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
        className={cn(
          fullScreen
            ? "max-w-[95vw] sm:max-w-[95vw] max-h-[95vh] w-[95vw]"
            : cn(desktopMaxWidth, "max-h-[90svh]"),
          "flex flex-col p-0 overflow-hidden gap-0",
          className
        )}
        onInteractOutside={
          dismissibleOnOutsideClick ? undefined : (e) => e.preventDefault()
        }
        onEscapeKeyDown={
          dismissibleOnOutsideClick ? undefined : (e) => e.preventDefault()
        }
      >
        {(title || description) && (
          <DialogHeader className="px-6 pt-5 pb-2 shrink-0">
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
        <div className="flex-1 overflow-y-auto px-6 pb-4 min-h-0">
          {children}
        </div>
        {footer && (
          <DialogFooter className="shrink-0 px-6 py-4 border-t bg-background">
            {footer}
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  );
}
