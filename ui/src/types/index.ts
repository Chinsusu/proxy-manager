export interface User {
  id: number;
  email: string;
  role: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  expires_in: number;
}

export interface Server {
  id: number;
  name: string;
  tags: string;
  wan_iface: string;
  lan_iface: string;
  last_seen_at: string | null;
  status: 'online' | 'offline';
  config_version: number;
  created_at: string;
  updated_at: string;
  proxies?: Proxy[];
  mappings?: Mapping[];
}

export interface CreateServerRequest {
  name: string;
  tags: string[];
  wan_iface: string;
  lan_iface: string;
}

export interface UpdateServerRequest {
  name?: string;
  tags?: string[];
  wan_iface?: string;
  lan_iface?: string;
  status?: string;
}

export interface Proxy {
  id: number;
  server_id: number;
  label: string;
  type: 'http' | 'https' | 'socks4' | 'socks5';
  host: string;
  port: number;
  username: string;
  password: string;
  health: 'ok' | 'fail' | 'unknown';
  created_at: string;
  updated_at: string;
  server?: Server;
}

export interface CreateProxyRequest {
  server_id: number;
  label: string;
  type: 'http' | 'https' | 'socks4' | 'socks5';
  host: string;
  port: number;
  username?: string;
  password?: string;
}

export interface UpdateProxyRequest {
  label?: string;
  type?: 'http' | 'https' | 'socks4' | 'socks5';
  host?: string;
  port?: number;
  username?: string;
  password?: string;
  health?: 'ok' | 'fail' | 'unknown';
}

export interface Mapping {
  id: number;
  server_id: number;
  client_cidr: string;
  dst_ports: string; // JSON array as string
  upstream_proxy_id: number;
  enabled: boolean;
  notes: string;
  created_at: string;
  updated_at: string;
  server?: Server;
  upstream_proxy?: Proxy;
}

export interface CreateMappingRequest {
  server_id: number;
  client_cidr: string;
  dst_ports: number[];
  upstream_proxy_id: number;
  enabled?: boolean;
  notes?: string;
}

export interface UpdateMappingRequest {
  client_cidr?: string;
  dst_ports?: number[];
  upstream_proxy_id?: number;
  enabled?: boolean;
  notes?: string;
}

export interface Summary {
  servers: number;
  proxies: number;
  mappings: number;
  active_servers: number;
  timestamp: string;
}
