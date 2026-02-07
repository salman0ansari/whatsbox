import { apiClient } from './client';
import type { 
  ConnectionStatus, 
  QRResponse, 
  Stats, 
  HourlyStats, 
  DailyStats,
  AuthResponse 
} from '@/types';

// Auth endpoints
export async function login(password: string): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>('/admin/login', { password });
  return data;
}

export async function logoutSession(): Promise<void> {
  await apiClient.post('/admin/logout-session');
}

export async function checkAuth(): Promise<{ authenticated: boolean }> {
  const { data } = await apiClient.get<{ authenticated: boolean }>('/admin/me');
  return data;
}

// Public status endpoint (no auth required)
export async function getPublicStatus(): Promise<{ connected: boolean }> {
  const { data } = await apiClient.get<{ connected: boolean }>('/status');
  return data;
}

// Admin status endpoints (auth required)
export async function getStatus(): Promise<ConnectionStatus> {
  const { data } = await apiClient.get<ConnectionStatus>('/admin/status');
  return data;
}

export async function getQRCode(): Promise<QRResponse> {
  const { data } = await apiClient.get<QRResponse>('/admin/qr');
  return data;
}

export async function logoutWhatsApp(): Promise<void> {
  await apiClient.post('/admin/logout');
}

// Stats endpoints
export async function getStats(): Promise<Stats> {
  const { data } = await apiClient.get<Stats>('/admin/stats');
  return data;
}

export async function getHourlyStats(hours = 24): Promise<HourlyStats[]> {
  const { data } = await apiClient.get<{ data: HourlyStats[] }>('/admin/stats/hourly', {
    params: { hours },
  });
  return data.data;
}

export async function getDailyStats(days = 30): Promise<DailyStats[]> {
  const { data } = await apiClient.get<{ data: DailyStats[] }>('/admin/stats/daily', {
    params: { days },
  });
  return data.data;
}
