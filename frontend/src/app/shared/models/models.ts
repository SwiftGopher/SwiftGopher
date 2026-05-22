export interface User {
  id: string;
  email: string;
  role: 'admin' | 'dispatcher' | 'courier' | 'client';
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
}

export interface Order {
  id: string;
  client_id: string;
  courier_id?: string;
  pickup_lat: number;
  pickup_lng: number;
  pickup_address: string;
  delivery_address: string;
  status: OrderStatus;
  price: number;
  created_at: string;
  updated_at: string;
}

export type OrderStatus = 'pending' | 'assigned' | 'in_progress' | 'delivered' | 'cancelled';

export interface OrderHistory {
  id: string;
  order_id: string;
  old_status: OrderStatus;
  new_status: OrderStatus;
  changed_at: string;
}

export interface OrderFilter {
  status?: OrderStatus | '';
  limit?: number;
  offset?: number;
  sort_by?: string;
  sort_dir?: 'asc' | 'desc';
}

export interface Courier {
  id: string;
  user_id: string;
  transport_type: 'bike' | 'car' | 'foot' | 'scooter';
  status: 'free' | 'busy' | 'offline';
  current_lat: number;
  current_lng: number;
}

export interface CreateOrderRequest {
  pickup_address: string;
  delivery_address: string;
  pickup_lat: number;
  pickup_lng: number;
  price: number;
}

export interface RegisterRequest {
  email: string;
  password: string;
  role: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}
