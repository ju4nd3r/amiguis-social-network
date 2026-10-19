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
        err.message || "Inicia sesión para exportar tus datos personales."
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
      <div className="bg-[#1E2229] border border-gray-800 rounded-2xl p-5">
        <div className="flex items-center space-x-2.5 mb-2">
          <ShieldCheck className="w-6 h-6 text-[#FCD116]" />
          <h1 className="text-lg font-extrabold text-white">
            Habeas Data Colombia
          </h1>
        </div>
        <p className="text-xs text-gray-300 leading-relaxed">
          En cumplimiento de la{" "}
          <strong className="text-white">
            Ley Estatutaria 1581 de 2012 y el Decreto 1377 de 2013
          </strong>
          , en Amiguis eres el único dueño de tu información. No vendemos datos,
          no creamos perfiles publicitarios ni usamos rastreadores de terceros.
        </p>

        <div className="mt-4 grid grid-cols-2 gap-2 text-xs">
          <div className="bg-gray-900/80 border border-gray-800 p-3 rounded-xl">
            <Lock className="w-4 h-4 text-[#FCD116] mb-1" />
            <strong className="block text-white">Argon2id</strong>
            <span className="text-gray-400">
              Contraseñas cifradas con el estándar más seguro.
            </span>
          </div>
          <div className="bg-gray-900/80 border border-gray-800 p-3 rounded-xl">
            <ShieldCheck className="w-4 h-4 text-emerald-400 mb-1" />
            <strong className="block text-white">Cero Rastreo</strong>
            <span className="text-gray-400">
              Sin píxeles espías ni telemetría comercial.
            </span>
          </div>
        </div>
      </div>

      {/* Right to Access / Portability */}
      <div className="bg-[#1E2229] border border-gray-800 rounded-2xl p-5">
        <h2 className="text-sm font-bold text-white mb-1">
          1. Derecho de Acceso y Portabilidad
        </h2>
        <p className="text-xs text-gray-400 mb-3">
          Descarga una copia completa en formato JSON de tus datos personales registrados en Amiguis.
        </p>
        <button
          onClick={handleExport}
          className="touch-target px-4 py-2 bg-[#003893] hover:bg-blue-800 text-white text-xs font-bold rounded-xl flex items-center space-x-2"
        >
          <Download className="w-4 h-4" />
          <span>Exportar mis datos personales (JSON)</span>
        </button>

        {exportedData && (
          <pre className="mt-3 p-3 bg-gray-950 border border-gray-800 rounded-xl text-[11px] text-emerald-400 overflow-x-auto">
            {exportedData}
          </pre>
        )}
      </div>

      {/* Right to Erasure */}
      <form
        onSubmit={handleDeleteAccount}
        className="bg-[#1E2229] border border-red-900/40 rounded-2xl p-5 space-y-3"
      >
        <h2 className="text-sm font-bold text-red-400">
          2. Derecho de Supresión (Eliminar Cuenta)
        </h2>
        <p className="text-xs text-gray-400">
          Solicita la eliminación total e irreversible de tu cuenta, perfil y publicaciones de nuestra base de datos.
        </p>

        <input
          type="password"
          required
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder="Confirma tu contraseña actual"
          className="w-full bg-gray-900 border border-gray-700 rounded-xl px-3 py-2 text-xs text-white"
        />

        {statusMessage && (
          <p className="text-xs text-amber-400">{statusMessage}</p>
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
