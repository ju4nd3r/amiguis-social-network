"use client";

import { useEffect, useState } from "react";
import { api, Profile, User } from "@/api/client";
import { MapPin, Globe, LogOut, Save } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";

export default function ProfilePage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<Profile | null>(null);
  const [displayName, setDisplayName] = useState("");
  const [bio, setBio] = useState("");
  const [city, setCity] = useState("");
  const [website, setWebsite] = useState("");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("amiguis_user");
      if (saved) {
        try {
          const parsed: User = JSON.parse(saved);
          setUser(parsed);
          setDisplayName(parsed.display_name || "");
          setCity(parsed.city || "Bogotá");
          api
            .getProfile(parsed.username)
            .then((p) => {
              setProfile(p);
              setDisplayName(p.display_name || "");
              setBio(p.bio || "");
              setCity(p.city || "");
              setWebsite(p.website || "");
            })
            .catch(() => {});
        } catch {}
      }
    }
  }, []);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setMessage(null);
    try {
      const updated = await api.updateProfile({
        display_name: displayName,
        bio,
        city,
        website,
      });
      setProfile(updated);
      setMessage("Perfil actualizado correctamente.");
    } catch (err: any) {
      setMessage(err.message || "Error al guardar el perfil");
    } finally {
      setSaving(false);
    }
  };

  const handleLogout = () => {
    api.clearTokens();
    router.push("/login");
  };

  if (!user) {
    return (
      <div className="p-4">
        <div className="bg-[#1E2229] border border-gray-800 rounded-2xl p-6 text-center">
          <h1 className="text-base font-bold text-white mb-2">
            Inicia sesión para ver tu perfil
          </h1>
          <p className="text-xs text-gray-400 mb-4">
            Personaliza tu biografía, ciudad en Colombia y círculos de amigos.
          </p>
          <Link
            href="/login"
            className="touch-target px-5 py-2.5 bg-[#FCD116] text-[#121417] font-bold text-xs rounded-xl"
          >
            Ir a Iniciar Sesión
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="p-4 space-y-4">
      <div className="bg-[#1E2229] border border-gray-800 rounded-2xl p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-3.5">
            <div className="w-14 h-14 rounded-full bg-gradient-to-tr from-[#003893] to-[#FCD116] flex items-center justify-center font-extrabold text-xl text-white">
              {displayName?.charAt(0).toUpperCase() || "A"}
            </div>
            <div>
              <h1 className="text-lg font-extrabold text-white">
                {displayName}
              </h1>
              <p className="text-xs text-gray-400">@{user.username}</p>
              {city && (
                <div className="flex items-center text-xs text-amber-400 mt-1">
                  <MapPin className="w-3.5 h-3.5 mr-1" />
                  <span>{city}, Colombia</span>
                </div>
              )}
            </div>
          </div>

          <button
            onClick={handleLogout}
            className="touch-target px-3 py-1.5 bg-gray-800 hover:bg-red-500/20 hover:text-red-400 text-xs text-gray-300 rounded-xl flex items-center space-x-1"
          >
            <LogOut className="w-3.5 h-3.5" />
            <span>Salir</span>
          </button>
        </div>

        <div className="grid grid-cols-3 gap-2 pt-3 border-t border-gray-800 text-center">
          <div>
            <span className="block text-sm font-extrabold text-white">
              {profile?.posts_count ?? 0}
            </span>
            <span className="text-[11px] text-gray-400">Publicaciones</span>
          </div>
          <div>
            <span className="block text-sm font-extrabold text-white">
              {profile?.followers_count ?? 0}
            </span>
            <span className="text-[11px] text-gray-400">Seguidores</span>
          </div>
          <div>
            <span className="block text-sm font-extrabold text-white">
              {profile?.following_count ?? 0}
            </span>
            <span className="text-[11px] text-gray-400">Siguiendo</span>
          </div>
        </div>
      </div>

      {/* Edit Profile Form */}
      <form
        onSubmit={handleSave}
        className="bg-[#1E2229] border border-gray-800 rounded-2xl p-5 space-y-3.5"
      >
        <h2 className="text-sm font-bold text-white">Editar mi perfil</h2>

        <div>
          <label className="block text-xs text-gray-400 mb-1">
            Nombre visible
          </label>
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className="w-full bg-gray-900 border border-gray-700 rounded-xl px-3 py-2 text-sm text-white"
          />
        </div>

        <div>
          <label className="block text-xs text-gray-400 mb-1">Biografía</label>
          <textarea
            rows={2}
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            placeholder="Cuéntale a tus parceros sobre ti..."
            className="w-full bg-gray-900 border border-gray-700 rounded-xl px-3 py-2 text-sm text-white resize-none"
          />
        </div>

        <div>
          <label className="block text-xs text-gray-400 mb-1">Ciudad</label>
          <input
            type="text"
            value={city}
            onChange={(e) => setCity(e.target.value)}
            placeholder="Medellín, Bogotá, Cali..."
            className="w-full bg-gray-900 border border-gray-700 rounded-xl px-3 py-2 text-sm text-white"
          />
        </div>

        <div>
          <label className="block text-xs text-gray-400 mb-1">
            Sitio web / enlace
          </label>
          <div className="flex items-center bg-gray-900 border border-gray-700 rounded-xl px-3 py-2">
            <Globe className="w-4 h-4 text-gray-500 mr-2" />
            <input
              type="text"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              placeholder="https://..."
              className="bg-transparent text-sm text-white w-full focus:outline-none"
            />
          </div>
        </div>

        {message && (
          <p className="text-xs text-[#FCD116] font-medium">{message}</p>
        )}

        <button
          type="submit"
          disabled={saving}
          className="w-full touch-target py-2.5 bg-[#003893] hover:bg-blue-800 text-white text-xs font-bold rounded-xl flex items-center justify-center space-x-1.5"
        >
          <Save className="w-4 h-4" />
          <span>{saving ? "Guardando..." : "Guardar cambios"}</span>
        </button>
      </form>
    </div>
  );
}
