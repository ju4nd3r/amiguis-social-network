"use client";

import { CreatePost } from "@/components/CreatePost";
import { useRouter } from "next/navigation";

export default function CreatePage() {
  const router = useRouter();

  return (
    <div className="p-4">
      <h1 className="text-lg font-extrabold text-gray-900 dark:text-white mb-2">
        Nueva publicación 🇨🇴
      </h1>
      <p className="text-xs text-gray-500 dark:text-gray-400 mb-4">
        Comparte una idea, foto o video con tus amiguis.
      </p>
      <CreatePost onPostCreated={() => router.push("/")} />
    </div>
  );
}
