"use client";

import { useState } from "react";
import { api } from "@/api/client";
import { ShieldCheck, Download, Trash2, Lock } from "lucide-react";
import { useRouter } from "next/navigation";

export default function PrivacyPage() {
  const router = useRouter();
  const [exportedData, setExportedData] = useState<string | null>(null);
  const [confirmPassword, setConfirmPassword] = useState("");
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const handleExport = async () => {
    setStatusMessage(null);
    try {
      const data = await api.exportData();
      setExportedData(JSON.stringify(data, null, 2));
    } catch (err: any) {
      setStatusMessage(
        err.message || "Inicia sesión para descargar tus datos personales."
      );
    }
  };

  const handleDeleteAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    if (
      !confirm(
        "¿Estás seguro? Esta acción eliminará permanentemente tu cuenta y todos tus datos conforme al derecho de supresión (Ley 1581 de 2012)."
      )
    ) {
      return;
    }

    try {
      await api.deleteAccount(confirmPassword);
      api.clearTokens();
      alert("Cuenta eliminada permanentemente.");
      router.push("/login");
    } catch (err: any) {
      setStatusMessage(err.message || "Error al eliminar la cuenta");
    }
  };

  return (
    <div className="p-4 space-y-4">
      <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-5 transition-colors">
        <div className="flex items-center space-x-2.5 mb-2">
          <ShieldCheck className="w-6 h-6 text-amber-500 dark:text-[#FCD116]" />
          <h1 className="text-lg font-extrabold text-gray-900 dark:text-white">
            Habeas Data Colombia
          </h1>
        </div>
        <p className="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
          En cumplimiento de la{" "}
          <strong className="text-gray-900 dark:text-white">
            Ley Estatutaria 1581 de 2012 y el Decreto 1377 de 2013
          </strong>
          , en Amiguis eres el único dueño de tu información. No vendemos tus
          datos ni compartimos tu información con terceros.
        </p>

        <div className="mt-4 grid grid-cols-2 gap-2 text-xs">
          <div className="bg-gray-50 dark:bg-gray-900/80 border border-gray-200 dark:border-gray-800 p-3 rounded-xl">
            <Lock className="w-4 h-4 text-amber-500 dark:text-[#FCD116] mb-1" />
            <strong className="block text-gray-900 dark:text-white">
              Seguridad Total
            </strong>
            <span className="text-gray-500 dark:text-gray-400">
              Tu cuenta y tu contraseña están siempre protegidas.
            </span>
          </div>
          <div className="bg-gray-50 dark:bg-gray-900/80 border border-gray-200 dark:border-gray-800 p-3 rounded-xl">
            <ShieldCheck className="w-4 h-4 text-emerald-500 dark:text-emerald-400 mb-1" />
            <strong className="block text-gray-900 dark:text-white">
              Privacidad Real
            </strong>
            <span className="text-gray-500 dark:text-gray-400">
              Tus conversaciones y publicaciones son solo para tus amiguis.
            </span>
          </div>
        </div>
      </div>

      {/* Right to Access / Portability */}
      <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-5 transition-colors">
        <h2 className="text-sm font-bold text-gray-900 dark:text-white mb-1">
          1. Conocer y Descargar mis Datos
        </h2>
        <p className="text-xs text-gray-500 dark:text-gray-400 mb-3">
          Descarga una copia completa de toda tu información registrada en Amiguis.
        </p>
        <button
          onClick={handleExport}
          className="touch-target px-4 py-2 bg-[#003893] hover:bg-blue-800 text-white text-xs font-bold rounded-xl flex items-center space-x-2"
        >
          <Download className="w-4 h-4" />
          <span>Descargar mis datos personales</span>
        </button>

        {exportedData && (
          <pre className="mt-3 p-3 bg-gray-900 border border-gray-800 rounded-xl text-[11px] text-emerald-400 overflow-x-auto">
            {exportedData}
          </pre>
        )}
      </div>

      {/* Right to Erasure */}
      <form
        onSubmit={handleDeleteAccount}
        className="bg-white dark:bg-[#1E2229] border border-red-300 dark:border-red-900/40 rounded-2xl p-5 space-y-3 transition-colors"
      >
        <h2 className="text-sm font-bold text-red-600 dark:text-red-400">
          2. Eliminar mi Cuenta
        </h2>
        <p className="text-xs text-gray-500 dark:text-gray-400">
          Solicita la eliminación total y definitiva de tu cuenta, perfil y publicaciones.
        </p>

        <input
          type="password"
          required
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder="Confirma tu contraseña actual"
          className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2 text-xs text-gray-900 dark:text-white"
        />

        {statusMessage && (
          <p className="text-xs text-amber-600 dark:text-amber-400">
            {statusMessage}
          </p>
        )}

        <button
          type="submit"
          className="touch-target px-4 py-2 bg-[#CE1126] hover:bg-red-700 text-white text-xs font-bold rounded-xl flex items-center space-x-2"
        >
          <Trash2 className="w-4 h-4" />
          <span>Eliminar mi cuenta permanentemente</span>
        </button>
      </form>
    </div>
  );
}
