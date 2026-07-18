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
}

export interface ListingImage {
  id: number;
  url: string;
  is_primary: boolean;
}

export interface ListingAttribute {
  field_key: string;
  label_ar: string;
  field_type: string;
  unit: string | null;
  value: unknown;
}

export interface ListingDetail {
  id: number;
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
