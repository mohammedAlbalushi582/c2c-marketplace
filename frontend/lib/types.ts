export interface User {
  id: number;
  email: string;
  username: string | null;
  full_name: string;
  phone: string | null;
  whatsapp_number: string | null;
  role: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface Category {
  id: number;
  parent_id: number | null;
  slug: string;
  name_ar: string;
  name_en: string | null;
  icon: string | null;
  display_order: number;
}

export interface FieldOption {
  value: string;
  label_ar: string;
  label_en?: string;
}

export interface CategoryField {
  id: number;
  field_key: string;
  label_ar: string;
  label_en: string | null;
  field_type: "text" | "textarea" | "number" | "select" | "multiselect" | "boolean" | "date" | "url";
  unit: string | null;
  options?: FieldOption[] | null;
  is_required: boolean;
  is_filterable: boolean;
  display_order: number;
}

export interface Location {
  id: number;
  parent_id: number | null;
  type: "governorate" | "wilayat";
  name_ar: string;
  name_en: string | null;
  slug: string;
}

export interface ListingCard {
  id: number;
  title: string;
  slug: string | null;
  price: number | null;
  currency: string;
  price_type: string;
  category_slug?: string;
  category_name_ar?: string;
  location_name_ar?: string | null;
  primary_image: string | null;
  status: string;
  is_featured: boolean;
  created_at: string;
  expires_at?: string | null;
}

export interface ListingImage {
  id: number;
  url: string;
  is_primary: boolean;
}

export interface ListingAttribute {
  field_id: number;
  field_key: string;
  label_ar: string;
  field_type: string;
  unit: string | null;
  value: unknown;
  /** Untranslated stored value — what the edit form writes back. */
  raw_value: unknown;
}

export interface ListingDetail {
  id: number;
  user_id: number;
  title: string;
  slug: string | null;
  description: string;
  price: number | null;
  currency: string;
  price_type: string;
  contact_phone: string | null;
  whatsapp_number: string | null;
  status: string;
  views_count: number;
  category_id: number;
  location_id: number | null;
  category_slug: string;
  category_name_ar: string;
  location_name_ar: string | null;
  created_at: string;
  images: ListingImage[];
  attributes: ListingAttribute[];
}

export interface SearchResponse {
  items: ListingCard[];
  total: number;
  page: number;
  page_size: number;
}

export interface AdminUser {
  id: number;
  email: string;
  username: string | null;
  full_name: string;
  phone: string | null;
  whatsapp_number: string | null;
  role: "user" | "admin";
  status: "active" | "suspended" | "deleted";
  listings_count: number;
  created_at: string;
}

export interface AdminUsersResponse {
  items: AdminUser[];
  total: number;
  page: number;
  page_size: number;
}

// ---- payments / fees ----

export interface FeeQuote {
  fee: number;
  currency: string;
  tier: number;
  live_count: number;
  free: boolean;
}

export interface PaymentInfo {
  id: number;
  amount: number;
  currency: string;
  status: string;
  purpose: string;
  checkout_url?: string;
}

/** Response of POST /listings and POST /listings/{id}/pay. */
export interface CreateListingResponse {
  listing: ListingDetail;
  payment?: PaymentInfo | null;
}

// ---- contact (راسلنا) ----

export interface ContactMessage {
  id: number;
  name: string;
  phone: string | null;
  email: string | null;
  message: string;
  is_read: boolean;
  created_at: string;
}

export interface ContactListResponse {
  items: ContactMessage[];
  total: number;
  unread: number;
  page: number;
  page_size: number;
}

export type AppSettings = Record<string, string>;

// ---- live presence (who's online) ----

export interface PresenceVisitor {
  name: string;
  is_user: boolean;
  path: string;
  seconds_ago: number;
}

export interface PresenceResponse {
  count: number;
  users: number;
  guests: number;
  visitors: PresenceVisitor[];
}
