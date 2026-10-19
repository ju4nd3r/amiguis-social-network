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

class ApiClient {
  private getToken(): string | null {
    if (typeof window !== "undefined") {
      return localStorage.getItem("amiguis_access_token");
    }
    return null;
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
      throw new Error(errData.error || `HTTP error ${res.status}`);
    }

    return res.json();
  }

  // Auth
  async register(data: { email: string; username: string; password: string; display_name: string; city?: string }) {
    const res = await this.request<{ access_token: string; refresh_token: string; user: User }>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
    this.setTokens(res.access_token, res.refresh_token);
    localStorage.setItem("amiguis_user", JSON.stringify(res.user));
    return res;
  }

  async login(loginStr: string, password: string) {
    const res = await this.request<{ access_token: string; refresh_token: string; user: User }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ login: loginStr, password }),
    });
    this.setTokens(res.access_token, res.refresh_token);
    localStorage.setItem("amiguis_user", JSON.stringify(res.user));
    return res;
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
    return this.request<Profile>("/profiles/me", {
      method: "PUT",
      body: JSON.stringify(data),
    });
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
