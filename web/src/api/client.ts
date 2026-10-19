// Typed API Client matching api/openapi.yaml single source of truth
import {
  CatalogCategory,
  CatalogItem,
  INITIAL_CATALOG_ITEMS,
  ReviewMetadata,
} from "@/data/catalog";

export type { CatalogCategory, CatalogItem, ReviewMetadata };

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
  type: "text" | "photo" | "video" | "review";
  body: string;
  media?: MediaItem[];
  review?: ReviewMetadata;
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
  public getLocalPosts(): Post[] {
    if (typeof window === "undefined") return [];
    try {
      const raw = localStorage.getItem("amiguis_local_posts");
      return raw ? JSON.parse(raw) : [];
    } catch {
      return [];
    }
  }

  public saveLocalPost(post: Post) {
    if (typeof window === "undefined") return;
    const current = this.getLocalPosts().filter((p) => p.id !== post.id);
    localStorage.setItem("amiguis_local_posts", JSON.stringify([post, ...current]));
  }

  public getFollowingIds(): string[] {
    if (typeof window === "undefined") return ["u-1", "u-2", "u-3"];
    try {
      const raw = localStorage.getItem("amiguis_following_ids");
      if (!raw) {
        const defaults = ["u-1", "u-2", "u-3"];
        localStorage.setItem("amiguis_following_ids", JSON.stringify(defaults));
        return defaults;
      }
      return JSON.parse(raw);
    } catch {
      return ["u-1", "u-2", "u-3"];
    }
  }

  public toggleFollowAuthor(authorId: string): boolean {
    if (typeof window === "undefined") return false;
    const current = new Set(this.getFollowingIds());
    let nowFollowing = false;
    if (current.has(authorId)) {
      current.delete(authorId);
      nowFollowing = false;
      this.unfollowUser(authorId).catch(() => {});
    } else {
      current.add(authorId);
      nowFollowing = true;
      this.followUser(authorId).catch(() => {});
    }
    localStorage.setItem("amiguis_following_ids", JSON.stringify(Array.from(current)));
    window.dispatchEvent(new Event("amiguis:follow-change"));
    return nowFollowing;
  }

  async createPost(
    type: "text" | "photo" | "video" | "review",
    body: string,
    media?: MediaItem[],
    review?: ReviewMetadata
  ): Promise<Post> {
    const post = await this.request<Post>("/posts", {
      method: "POST",
      body: JSON.stringify({ type, body, media, review }),
    });
    this.saveLocalPost(post);
    return post;
  }

  async getPost(id: string): Promise<Post> {
    return this.request<Post>(`/posts/${id}`);
  }

  async deletePost(id: string): Promise<{ success: boolean }> {
    if (typeof window !== "undefined") {
      const next = this.getLocalPosts().filter((p) => p.id !== id);
      localStorage.setItem("amiguis_local_posts", JSON.stringify(next));
    }
    return this.request<{ success: boolean }>(`/posts/${id}`, { method: "DELETE" });
  }

  // Catalog (Movies, Series & Books) & Reviews
  private getLocalCatalog(): CatalogItem[] {
    if (typeof window === "undefined") return INITIAL_CATALOG_ITEMS;
    try {
      const raw = localStorage.getItem("amiguis_catalog_items");
      if (!raw) {
        localStorage.setItem("amiguis_catalog_items", JSON.stringify(INITIAL_CATALOG_ITEMS));
        return INITIAL_CATALOG_ITEMS;
      }
      return JSON.parse(raw);
    } catch {
      return INITIAL_CATALOG_ITEMS;
    }
  }

  private saveLocalCatalog(items: CatalogItem[]) {
    if (typeof window !== "undefined") {
      localStorage.setItem("amiguis_catalog_items", JSON.stringify(items));
    }
  }

  async searchCatalog(query = "", category = "all"): Promise<CatalogItem[]> {
    try {
      const params = new URLSearchParams();
      if (query) params.set("q", query);
      if (category && category !== "all") params.set("category", category);
      const res = await this.request<{ items: CatalogItem[] }>(`/catalog?${params.toString()}`);
      if (res && Array.isArray(res.items)) {
        return res.items;
      }
    } catch {
      // Fallback to local catalog store
    }
    const items = this.getLocalCatalog();
    const q = query.trim().toLowerCase();
    const cat = category.trim().toLowerCase();
    return items.filter((item) => {
      if (cat && cat !== "all" && item.category !== cat) return false;
      if (!q) return true;
      return (
        item.title.toLowerCase().includes(q) ||
        item.creator.toLowerCase().includes(q) ||
        item.producer.toLowerCase().includes(q) ||
        item.cast.some((c) => c.toLowerCase().includes(q)) ||
        item.genres.some((g) => g.toLowerCase().includes(q))
      );
    });
  }

  async getCatalogItem(id: string): Promise<CatalogItem | null> {
    try {
      const item = await this.request<CatalogItem>(`/catalog/${id}`);
      if (item && item.id) return item;
    } catch {
      // Fallback to local catalog store
    }
    const items = this.getLocalCatalog();
    return items.find((i) => i.id === id) || null;
  }

  async submitCatalogReview(
    itemId: string,
    data: { body: string; rating: number; watched: boolean; liked: boolean }
  ): Promise<{ post: Post; item: CatalogItem }> {
    let currentAuthor: Profile = {
      id: "local-user",
      username: "amigui_col",
      display_name: "Amigui",
      city: "Bogotá",
      followers_count: 1,
      following_count: 3,
      posts_count: 1,
      created_at: new Date().toISOString(),
    };

    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("amiguis_user");
      if (saved) {
        try {
          const parsed: User = JSON.parse(saved);
          currentAuthor = {
            id: parsed.id || "local-user",
            username: parsed.username || "amigui_col",
            display_name: parsed.display_name || parsed.username || "Amigui",
            city: parsed.city || "Bogotá",
            followers_count: 1,
            following_count: 3,
            posts_count: 1,
            created_at: parsed.created_at || new Date().toISOString(),
          };
        } catch {}
      }
    }

    // Update local catalog item rating as well
    const catalog = this.getLocalCatalog();
    const idx = catalog.findIndex((i) => i.id === itemId);
    const baseItem = idx >= 0 ? catalog[idx] : INITIAL_CATALOG_ITEMS[0];
    const updatedItem: CatalogItem = { ...baseItem };
    if (data.rating > 0) {
      const total = updatedItem.average_rating * updatedItem.ratings_count + data.rating;
      updatedItem.ratings_count += 1;
      updatedItem.average_rating = Math.round((total / updatedItem.ratings_count) * 10) / 10;
      if (idx >= 0) {
        catalog[idx] = updatedItem;
        this.saveLocalCatalog(catalog);
      }
    }

    try {
      const res = await this.request<{ post: Post; item: CatalogItem }>(`/catalog/${itemId}/review`, {
        method: "POST",
        body: JSON.stringify(data),
      });
      if (res && res.post) {
        const enrichedPost: Post = {
          ...res.post,
          author:
            res.post.author?.username && res.post.author.username !== "amigui"
              ? res.post.author
              : currentAuthor,
        };
        this.saveLocalPost(enrichedPost);
        return { post: enrichedPost, item: res.item || updatedItem };
      }
    } catch {
      // Fallback to local post creation when backend is offline
    }

    const reviewMeta: ReviewMetadata = {
      item_id: updatedItem.id,
      category: updatedItem.category,
      item_title: updatedItem.title,
      item_year: updatedItem.year,
      poster_url: updatedItem.poster_url,
      rating: data.rating,
      watched: data.watched,
      liked: data.liked,
      creator: updatedItem.creator,
      producer: updatedItem.producer,
    };

    const cleanBody =
      data.body.trim() ||
      (data.watched
        ? `Marcó "${updatedItem.title}" como vista/visto.`
        : `Compartió una reseña sobre "${updatedItem.title}".`);

    const localPost: Post = {
      id: `review-${Date.now()}`,
      author: currentAuthor,
      type: "review",
      body: cleanBody,
      review: reviewMeta,
      reactions_count: 0,
      comments_count: 0,
      comments: [],
      created_at: new Date().toISOString(),
    };

    this.saveLocalPost(localPost);
    return { post: localPost, item: updatedItem };
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
