"use client";

import { useEffect, useRef, useState } from "react";
import { api } from "@/api/client";
import { useRouter } from "next/navigation";
import { CheckCircle2, XCircle } from "lucide-react";

const COLOMBIAN_CITIES = [
  "Bogotá",
  "Medellín",
  "Cali",
  "Barranquilla",
  "Bucaramanga",
  "Cartagena",
  "Pereira",
  "Manizales",
  "Santa Marta",
  "Ibagué",
  "Pasto",
  "Cúcuta",
  "Villavicencio",
  "Armenia",
  "Popayán",
];

interface StatusModalState {
  status: "success" | "error";
  title: string;
  message: string;
}

export default function LoginPage() {
  const router = useRouter();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [city, setCity] = useState("Bogotá");
  const [password, setPassword] = useState("");
  const [acceptedHabeasData, setAcceptedHabeasData] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [statusModal, setStatusModal] = useState<StatusModalState | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  const showTempModal = (modal: StatusModalState, redirectOnClose = false) => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }
    setStatusModal(modal);
    const duration = modal.status === "success" ? 2000 : 4000;
    timerRef.current = setTimeout(() => {
      setStatusModal(null);
      if (redirectOnClose) {
        router.push("/");
      }
    }, duration);
  };

  const handleCloseModal = () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }
    const wasSuccess = statusModal?.status === "success";
    setStatusModal(null);
    if (wasSuccess) {
      router.push("/");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (mode === "register" && !acceptedHabeasData) {
      const msg =
        "Debes autorizar el tratamiento de datos personales conforme a la Ley 1581 de 2012.";
      setError(msg);
      showTempModal({
        status: "error",
        title: "No se pudo completar el registro",
        message: msg,
      });
      return;
    }

    setLoading(true);
    try {
      if (mode === "login") {
        const res = await api.login(email || username, password);
        const name = res.user.display_name || res.user.username;
        showTempModal(
          {
            status: "success",
            title: "¡Sesión iniciada!",
            message: `¡Hola de nuevo, ${name}! Ya estás conectado con tus amiguis.`,
          },
          true
        );
      } else {
        const res = await api.register({
          email,
          username,
          password,
          display_name: displayName || username,
          city,
        });
        const name = res.user.display_name || res.user.username;
        showTempModal(
          {
            status: "success",
            title: "¡Cuenta creada con éxito!",
            message: `¡Bienvenido a Amiguis, ${name}! Tu registro se realizó correctamente.`,
          },
          true
        );
      }
    } catch (err: any) {
      const msg = err.message || "Error de autenticación";
      setError(msg);
      showTempModal({
        status: "error",
        title:
          mode === "register"
            ? "No se pudo completar el registro"
            : "Error al iniciar sesión",
        message: msg,
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-4">
      {/* Temporary Result Modal */}
      {statusModal && (
        <div
          role="dialog"
          aria-modal="true"
          data-testid="auth-status-modal"
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-150"
        >
          <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-700 rounded-2xl p-6 max-w-sm w-full shadow-2xl text-center space-y-4">
            <div className="flex justify-center">
              {statusModal.status === "success" ? (
                <div className="w-14 h-14 rounded-full bg-emerald-500/15 text-emerald-500 flex items-center justify-center">
                  <CheckCircle2 className="w-8 h-8" />
                </div>
              ) : (
                <div className="w-14 h-14 rounded-full bg-red-500/15 text-red-500 flex items-center justify-center">
                  <XCircle className="w-8 h-8" />
                </div>
              )}
            </div>

            <div className="space-y-1.5">
              <h2 className="text-base font-extrabold text-gray-900 dark:text-white">
                {statusModal.title}
              </h2>
              <p className="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
                {statusModal.message}
              </p>
            </div>

            <button
              type="button"
              onClick={handleCloseModal}
              data-testid="auth-modal-close"
              className={`w-full touch-target py-2.5 text-xs font-extrabold rounded-xl transition-colors ${
                statusModal.status === "success"
                  ? "bg-[#FCD116] text-[#121417] hover:bg-amber-300"
                  : "bg-gray-200 dark:bg-gray-800 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-700"
              }`}
            >
              {statusModal.status === "success" ? "Continuar" : "Entendido"}
            </button>
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-6 shadow-md transition-colors">
        <h1 className="text-xl font-extrabold text-gray-900 dark:text-white mb-1">
          {mode === "login"
            ? "Iniciar sesión en Amiguis"
            : "Crear cuenta en Amiguis 🇨🇴"}
        </h1>
        <p className="text-xs text-gray-500 dark:text-gray-400 mb-6">
          Conecta y comparte momentos reales con tus amiguis.
        </p>

        <div className="flex space-x-2 mb-6 bg-gray-100 dark:bg-gray-900 p-1 rounded-xl">
          <button
            type="button"
            onClick={() => {
              setMode("login");
              setError(null);
            }}
            className={`flex-1 touch-target py-2 text-xs font-bold rounded-lg transition-all ${
              mode === "login"
                ? "bg-[#FCD116] text-[#121417]"
                : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            }`}
          >
            Entrar
          </button>
          <button
            type="button"
            onClick={() => {
              setMode("register");
              setError(null);
            }}
            className={`flex-1 touch-target py-2 text-xs font-bold rounded-lg transition-all ${
              mode === "register"
                ? "bg-[#FCD116] text-[#121417]"
                : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            }`}
          >
            Registrarse
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {mode === "register" && (
            <>
              <div>
                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Correo electrónico
                </label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="camila@ejemplo.co"
                  className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2.5 text-sm text-gray-900 dark:text-white focus:outline-none focus:border-[#FCD116]"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Nombre visible
                </label>
                <input
                  type="text"
                  required
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Camila Restrepo"
                  className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2.5 text-sm text-gray-900 dark:text-white focus:outline-none focus:border-[#FCD116]"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Ciudad en Colombia
                </label>
                <select
                  value={city}
                  onChange={(e) => setCity(e.target.value)}
                  className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2.5 text-sm text-gray-900 dark:text-white focus:outline-none focus:border-[#FCD116]"
                >
                  {COLOMBIAN_CITIES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
              </div>
            </>
          )}

          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
              {mode === "login" ? "Usuario o correo" : "Nombre de usuario"}
            </label>
            <input
              type="text"
              required
              value={username}
              onChange={(e) => {
                setUsername(e.target.value);
                if (mode === "login") setEmail(e.target.value);
              }}
              placeholder="camila_medallo"
              className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2.5 text-sm text-gray-900 dark:text-white focus:outline-none focus:border-[#FCD116]"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
              Contraseña (mínimo 8 caracteres)
            </label>
            <input
              type="password"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2.5 text-sm text-gray-900 dark:text-white focus:outline-none focus:border-[#FCD116]"
            />
          </div>

          {mode === "register" && (
            <label className="flex items-start space-x-2.5 text-xs text-gray-700 dark:text-gray-300 cursor-pointer pt-1">
              <input
                type="checkbox"
                checked={acceptedHabeasData}
                onChange={(e) => setAcceptedHabeasData(e.target.checked)}
                className="mt-0.5 rounded border-gray-400 dark:border-gray-600 text-[#FCD116] focus:ring-[#FCD116]"
              />
              <span>
                Autorizo de manera libre y expresa el tratamiento de mis datos
                personales conforme a la{" "}
                <strong className="text-amber-600 dark:text-[#FCD116]">
                  Ley 1581 de 2012 (Habeas Data Colombia)
                </strong>
                .
              </span>
            </label>
          )}

          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/30 rounded-xl text-xs text-red-500 dark:text-red-400">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full touch-target py-3 bg-[#FCD116] text-[#121417] font-extrabold text-sm rounded-xl hover:bg-amber-300 transition-colors disabled:opacity-50"
          >
            {loading
              ? "Procesando..."
              : mode === "login"
              ? "Entrar a Amiguis"
              : "Crear mi cuenta"}
          </button>
        </form>
      </div>
    </div>
  );
}
