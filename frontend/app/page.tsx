"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/lib/auth-store";

export default function Home() {
  const router = useRouter();
  const { user, initialized, bootstrap } = useAuthStore();

  console.log("[Home] render — initialized:", initialized, "user:", user?.email);

  useEffect(() => {
    console.log("[Home] useEffect — calling bootstrap");
    bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    console.log("[Home] redirect effect — initialized:", initialized, "user:", user);
    if (initialized) {
      const target = user ? "/dashboard" : "/login";
      console.log("[Home] redirecting to:", target);
      router.push(target);
    }
  }, [initialized, user, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-950 text-white">
      <div className="text-zinc-400">
        Loading... (initialized: {String(initialized)})
      </div>
    </div>
  );
}
