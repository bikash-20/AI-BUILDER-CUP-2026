"use client";

import { ChatWindow } from "@/components/customer/ChatWindow";

export default function CustomerChatPage() {
  return (
    <section className="max-w-3xl">
      <h1 className="text-2xl font-semibold mb-4">FinGuard Assistant</h1>
      <ChatWindow />
    </section>
  );
}
