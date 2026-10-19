// Typed API Client matching api/openapi.yaml single source of truth

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  city?: string;
  is_verified?: boolean;
  created_at: string;
}

export interface Profile {
  id: string;
  username: string;
  display_name: string;
  bio?: string;
  avatar_url?: string;
  city?: string;
  website?: string;
  followers_count: number;
  following_count: number;
  posts_count: number;
  is_following?: boolean;
  is_muted?: boolean;
  is_blocked?: boolean;
  created_at: string;
}

export interface MediaItem {
  id: string;
  storage_key: string;
  url: string;
  thumbnail_url?: string;
  kind: "photo" | "video";
  status: "pending" | "processing" | "ready" | "failed";
}

export interface Post {
  id: string;
  author: Profile;
  type: "text" | "photo" | "video";
  body: string;
  media?: MediaItem[];
  reactions_count: number;
  comments_count: number;
  comments?: Comment[];
  my_reaction?: string;
  created_at: string;
}

export interface Comment {
  id: string;
  post_id: string;
  author: Profile;
  parent_id?: string;
  body: string;
  created_at: string;
  replies?: Comment[];
}

export interface FeedResponse {
  items: Post[];
  next_cursor?: string;
  has_more: boolean;
}

export interface Notification {
  id: string;
  type: "follow" | "reaction" | "comment" | "reply" | "mention";
  actor: Profile;
  entity_ref?: string;
  read_at?: string;
  created_at: string;
}

interface StoredLocalUser extends User {
  password: string;
}

function translateErrorMessage(msg: string): string {
  const lower = msg.toLowerCase();
  if (lower.includes("invalid credentials")) {
    return "Credenciales inválidas. Verifica tu usuario o contraseña.";
  }
  if (lower.includes("email already") || lower.includes("email exists")) {
    return "Este correo electrónico ya está registrado.";
  }
  if (lower.includes("username already") || lower.includes("username exists")) {
    return "Este nombre de usuario ya está en uso.";
  }
  if (lower.includes("password") && lower.includes("8")) {
    return "La contraseña debe tener al menos 8 caracteres.";
  }
  return msg;
}

class ApiClient {
  private getToken(): string | null {
    if (typeof window !== "undefined") {
      return localStorage.getItem("amiguis_access_token");
    }
    return null;
  }

  private notifyAuthChange() {
    if (typeof window !== "undefined") {
      window.dispatchEvent(new Event("amiguis:auth-change"));
    }
  }

  private getLocalUsers(): StoredLocalUser[] {
    if (typeof window === "undefined") return [];
    try {
      const raw = localStorage.getItem("amiguis_local_users");
      return raw ? JSON.parse(raw) : [];
    } catch {
      return [];
    }
  }

  private saveLocalUsers(users: StoredLocalUser[]) {
    if (typeof window !== "undefined") {
      localStorage.setItem("amiguis_local_users", JSON.stringify(users));
    }
  }

  public setTokens(access: string, refresh: string) {
    if (typeof window !== "undefined") {
      localStorage.setItem("amiguis_access_token", access);
      localStorage.setItem("amiguis_refresh_token", refresh);
    }
  }

  public clearTokens() {
    if (typeof window !== "undefined") {
      localStorage.removeItem("amiguis_access_token");
      localStorage.removeItem("amiguis_refresh_token");
      localStorage.removeItem("amiguis_user");
      this.notifyAuthChange();
    }
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({ error: "Error desconocido" }));
      throw new Error(translateErrorMessage(errData.error || `HTTP error ${res.status}`));
    }

    return res.json();
  }

  // Auth
  async register(data: { email: string; username: string; password: string; display_name: string; city?: string }) {
    try {
      const res = await this.request<{ access_token: string; refresh_token: string; user: User }>("/auth/register", {
        method: "POST",
        body: JSON.stringify(data),
      });
      this.setTokens(res.access_token, res.refresh_token);
      localStorage.setItem("amiguis_user", JSON.stringify(res.user));
      this.notifyAuthChange();
      return res;
    } catch (err: any) {
      // If the error came from the server (not a network offline TypeError), rethrow it
      if (!(err instanceof TypeError)) {
        throw err;
      }
      // Validated local fallback when running frontend standalone
      if (!data.password || data.password.length < 8) {
        throw new Error("La contraseña debe tener al menos 8 caracteres.");
      }
      const localUsers = this.getLocalUsers();
      if (localUsers.some((u) => u.email.toLowerCase() === data.email.toLowerCase())) {
        throw new Error("Este correo electrónico ya está registrado.");
      }
      if (localUsers.some((u) => u.username.toLowerCase() === data.username.toLowerCase())) {
        throw new Error("Este nombre de usuario ya está en uso.");
      }
      const newUser: User = {
        id: `u-${Date.now()}`,
        email: data.email.trim(),
        username: data.username.trim(),
        display_name: (data.display_name || data.username).trim(),
        city: data.city || "Bogotá",
        is_verified: true,
        created_at: new Date().toISOString(),
      };
      this.saveLocalUsers([...localUsers, { ...newUser, password: data.password }]);
      const res = {
        access_token: `local-access-${newUser.id}`,
        refresh_token: `local-refresh-${newUser.id}`,
        user: newUser,
      };
      this.setTokens(res.access_token, res.refresh_token);
      localStorage.setItem("amiguis_user", JSON.stringify(res.user));
      this.notifyAuthChange();
      return res;
    }
  }

  async login(loginStr: string, password: string) {
    try {
      const res = await this.request<{ access_token: string; refresh_token: string; user: User }>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ login: loginStr, password }),
      });
      this.setTokens(res.access_token, res.refresh_token);
      localStorage.setItem("amiguis_user", JSON.stringify(res.user));
      this.notifyAuthChange();
      return res;
    } catch (err: any) {
      if (!(err instanceof TypeError)) {
        throw err;
      }
      const cleanLogin = loginStr.trim().toLowerCase();
      const localUsers = this.getLocalUsers();
      const matched = localUsers.find(
        (u) =>
          (u.email.toLowerCase() === cleanLogin || u.username.toLowerCase() === cleanLogin) &&
          u.password === password
      );
      if (!matched) {
        throw new Error("Credenciales inválidas. Verifica tu usuario o contraseña.");
      }
      const { password: _, ...userWithoutPassword } = matched;
      const res = {
        access_token: `local-access-${userWithoutPassword.id}`,
        refresh_token: `local-refresh-${userWithoutPassword.id}`,
        user: userWithoutPassword,
      };
      this.setTokens(res.access_token, res.refresh_token);
      localStorage.setItem("amiguis_user", JSON.stringify(res.user));
      this.notifyAuthChange();
      return res;
    }
  }

  async getMe(): Promise<User> {
    return this.request<User>("/auth/me");
  }

  // Feed (Guaranteed strictly chronological)
  async getFeed(cursor = "", type = "all", listId = ""): Promise<FeedResponse> {
    const params = new URLSearchParams();
    if (cursor) params.set("cursor", cursor);
    if (type && type !== "all") params.set("type", type);
    if (listId) params.set("list_id", listId);
    return this.request<FeedResponse>(`/feed?${params.toString()}`);
  }

  // Posts
  async createPost(type: "text" | "photo" | "video", body: string, media?: MediaItem[]): Promise<Post> {
    return this.request<Post>("/posts", {
      method: "POST",
      body: JSON.stringify({ type, body, media }),
    });
  }

  async getPost(id: string): Promise<Post> {
    return this.request<Post>(`/posts/${id}`);
  }

  async deletePost(id: string): Promise<{ success: boolean }> {
    return this.request<{ success: boolean }>(`/posts/${id}`, { method: "DELETE" });
  }

  // Engagement
  async addReaction(postId: string, type: "interesante" = "interesante") {
    return this.request<{ success: boolean }>(`/posts/${postId}/reactions`, {
      method: "POST",
      body: JSON.stringify({ type }),
    });
  }

  async removeReaction(postId: string) {
    return this.request<{ success: boolean }>(`/posts/${postId}/reactions`, {
      method: "DELETE",
    });
  }

  async getComments(postId: string): Promise<Comment[]> {
    try {
      const res = await this.request<Comment[] | null>(`/posts/${postId}/comments`);
      return Array.isArray(res) ? res : [];
    } catch {
      return [];
    }
  }

  async addComment(postId: string, body: string, parentId?: string): Promise<Comment> {
    return this.request<Comment>(`/posts/${postId}/comments`, {
      method: "POST",
      body: JSON.stringify({ body, parent_id: parentId }),
    });
  }

  // Profiles & Social Graph
  async getProfile(username: string): Promise<Profile> {
    return this.request<Profile>(`/profiles/${username}`);
  }

  async updateProfile(data: { display_name?: string; bio?: string; city?: string; website?: string; avatar_url?: string }): Promise<Profile> {
    const updated = await this.request<Profile>("/profiles/me", {
      method: "PUT",
      body: JSON.stringify(data),
    });
    if (typeof window !== "undefined") {
      try {
        const raw = localStorage.getItem("amiguis_user");
        if (raw) {
          const parsed = JSON.parse(raw);
          const nextUser = {
            ...parsed,
            display_name: updated.display_name || data.display_name || parsed.display_name,
            city: updated.city || data.city || parsed.city,
            avatar_url: updated.avatar_url || data.avatar_url || parsed.avatar_url,
          };
          localStorage.setItem("amiguis_user", JSON.stringify(nextUser));
          this.notifyAuthChange();
        }
      } catch {}
    }
    return updated;
  }

  async followUser(userId: string) {
    return this.request<{ success: boolean }>(`/users/${userId}/follow`, { method: "PUT" });
  }

  async unfollowUser(userId: string) {
    return this.request<{ success: boolean }>(`/users/${userId}/follow`, { method: "DELETE" });
  }

  async muteUser(userId: string) {
    return this.request<{ success: boolean }>(`/users/${userId}/mute`, { method: "PUT" });
  }

  // Notifications
  async getNotifications(): Promise<Notification[]> {
    return this.request<Notification[]>("/notifications");
  }

  async getUnreadCount(): Promise<{ count: number }> {
    return this.request<{ count: number }>("/notifications/unread-count");
  }

  async markAllRead() {
    return this.request<{ success: boolean }>("/notifications/read-all", { method: "PUT" });
  }

  // Colombian Habeas Data Compliance (Ley 1581 de 2012)
  async exportData() {
    return this.request<any>("/privacy/export");
  }

  async deleteAccount(confirmPassword: string) {
    return this.request<{ success: boolean }>("/privacy/delete-account", {
      method: "DELETE",
      body: JSON.stringify({ confirm_password: confirmPassword }),
    });
  }
}

export const api = new ApiClient();
