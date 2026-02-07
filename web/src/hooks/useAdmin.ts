import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import {
  login,
  logoutSession,
  checkAuth,
  getPublicStatus,
  getStatus,
  getQRCode,
  logoutWhatsApp,
  getStats,
  getHourlyStats,
  getDailyStats,
} from '@/api/admin';
import { getErrorMessage } from '@/api/client';

// Auth hooks
export function useAuth() {
  return useQuery({
    queryKey: ['auth'],
    queryFn: checkAuth,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: login,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth'] });
      toast.success('Logged in successfully');
      navigate({ to: '/admin' });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
}

export function useLogoutSession() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  return useMutation({
    mutationFn: logoutSession,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth'] });
      toast.success('Logged out successfully');
      navigate({ to: '/admin/login' });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
}

// Public status hook (no auth required, for public header)
export function usePublicStatus() {
  return useQuery({
    queryKey: ['publicStatus'],
    queryFn: getPublicStatus,
    refetchInterval: 30000, // Refetch every 30 seconds
    retry: 1,
  });
}

// Admin status hooks (auth required)
export function useStatus() {
  return useQuery({
    queryKey: ['status'],
    queryFn: getStatus,
    refetchInterval: 10000, // Refetch every 10 seconds
  });
}

export function useQRCode() {
  const { data: status } = useStatus();

  return useQuery({
    queryKey: ['qr'],
    queryFn: getQRCode,
    enabled: status?.logged_in === false,
    refetchInterval: 20000, // Refetch every 20 seconds while waiting for scan
  });
}

export function useLogoutWhatsApp() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: logoutWhatsApp,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status'] });
      queryClient.invalidateQueries({ queryKey: ['qr'] });
      toast.success('Logged out from WhatsApp');
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
}

// Stats hooks
export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: getStats,
    refetchInterval: 30000, // Refetch every 30 seconds
  });
}

export function useHourlyStats(hours = 24) {
  return useQuery({
    queryKey: ['stats', 'hourly', hours],
    queryFn: () => getHourlyStats(hours),
    refetchInterval: 60000, // Refetch every minute
  });
}

export function useDailyStats(days = 30) {
  return useQuery({
    queryKey: ['stats', 'daily', days],
    queryFn: () => getDailyStats(days),
    refetchInterval: 5 * 60 * 1000, // Refetch every 5 minutes
  });
}
