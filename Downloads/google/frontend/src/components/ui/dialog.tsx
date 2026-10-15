"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { cn } from "@/lib/format";
import * as React from "react";

export const Modal = Dialog.Root;
export const ModalTrigger = Dialog.Trigger;
export const ModalClose = Dialog.Close;

export const ModalContent = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <Dialog.Portal>
    <Dialog.Overlay className="fixed inset-0 bg-black/40 z-40" />
    <Dialog.Content
      ref={ref}
      className={cn(
        "fixed left-1/2 top-1/2 z-50 w-[min(90vw,480px)] -translate-x-1/2 -translate-y-1/2 rounded-lg bg-white shadow-xl",
        className
      )}
      {...props}
    />
  </Dialog.Portal>
));
ModalContent.displayName = "ModalContent";
