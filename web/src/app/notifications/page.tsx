"use client";

import { useEffect, useState } from "react";
import { api, Notification } from "@/api/client";
import { Bell, CheckCheck } from "lucide-react";

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);

  const loadNotifications = async () => {
    setLoading(true);
    try {
      const items = await api.getNotifications();
      setNotifications(items || []);
    } catch {
      setNotifications([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadNotifications();
  }, []);

  const handleMarkAllRead = async () => {
    try {
      await api.markAllRead();
      loadNotifications();
    } catch {}
  };

  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          <Bell className="w-5 h-5 text-[#FCD116]" />
          <h1 className="text-lg font-extrabold text-white">Avisos</h1>
        </div>
        <button
          onClick={handleMarkAllRead}
          className="touch-target px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-xs font-semibold text-gray-300 rounded-xl flex items-center space-x-1.5"
        >
          <CheckCheck className="w-3.5 h-3.5" />
          <span>Marcar leídos</span>
        </button>
      </div>

      <p className="text-xs text-gray-400 mb-4">
        Solo recibes notificaciones reales de tus interacciones. Nada de alertas falsas para retener tu atención.
      </p>

      {loading ? (
        <div className="space-y-3">
          {[1, 2].map((i) => (
            <div
              key={i}
              className="bg-[#1E2229] h-16 rounded-2xl animate-pulse"
            />
          ))}
        </div>
      ) : notifications.length === 0 ? (
        <div className="bg-[#1E2229] border border-gray-800 rounded-2xl p-8 text-center">
          <p className="text-sm text-gray-400">
            Todo al día. No tienes notificaciones pendientes.
          </p>
        </div>
      ) : (
        <div className="space-y-2.5">
          {notifications.map((n) => (
            <div
              key={n.id}
              className={`p-3.5 rounded-2xl border flex items-center justify-between ${
                n.read_at
                  ? "bg-[#1E2229]/60 border-gray-800/60"
                  : "bg-[#1E2229] border-[#FCD116]/40"
              }`}
            >
              <div className="text-xs text-gray-200">
                <span className="font-bold text-white">
                  @{n.actor?.username || "alguien"}
                </span>{" "}
                {n.type === "follow" && "comenzó a seguirte."}
                {n.type === "reaction" && "reaccionó a tu publicación."}
                {n.type === "comment" && "comentó tu publicación."}
                {n.type === "reply" && "respondió a tu comentario."}
                {n.type === "mention" && "te mencionó."}
              </div>
              <span className="text-[10px] text-gray-500 ml-2">
                {new Date(n.created_at).toLocaleTimeString("es-CO", {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
