import { useMutation, useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { User, LoginRequest, LoginResponse } from '../types';
import toast from 'react-hot-toast';

export const useAuth = () => {
  const token = localStorage.getItem('auth_token');
  const isAuthenticated = !!token;

  const login = useMutation({
    mutationFn: async (data: LoginRequest): Promise<LoginResponse> => {
      const response = await api.post('/auth/login', data);
      return response.data;
    },
    onSuccess: (data) => {
      localStorage.setItem('auth_token', data.access_token);
      toast.success('Logged in successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Login failed');
    },
  });

  const logout = () => {
    localStorage.removeItem('auth_token');
    window.location.href = '/login';
  };

  const { data: user } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async (): Promise<User> => {
      const response = await api.get('/auth/me');
      return response.data;
    },
    enabled: isAuthenticated,
  });

  return {
    isAuthenticated,
    user,
    login,
    logout,
  };
};
