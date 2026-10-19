"use client";

import { CreatePost } from "@/components/CreatePost";
import { useRouter } from "next/navigation";

export default function CreatePage() {
  const router = useRouter();

  return (
    <div className="p-4">
      <h1 className="text-lg font-extrabold text-white mb-2">
        Nueva publicación 🇨🇴
      </h1>
      <p className="text-xs text-gray-400 mb-4">
        Tus seguidores la verán exactamente en el orden en que la publiques, sin algoritmos intermedios.
      </p>
      <CreatePost onPostCreated={() => router.push("/")} />
    </div>
  );
}
