import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Server, CreateServerRequest, UpdateServerRequest, Summary } from '../types';
import toast from 'react-hot-toast';

export const useServers = () => {
  const queryClient = useQueryClient();

  const { data: servers, isLoading, error } = useQuery({
    queryKey: ['servers'],
    queryFn: async (): Promise<Server[]> => {
      const response = await api.get('/servers');
      return response.data;
    },
  });

  const { data: summary } = useQuery({
    queryKey: ['admin', 'summary'],
    queryFn: async (): Promise<Summary> => {
      const response = await api.get('/admin/summary');
      return response.data;
    },
  });

  const createServer = useMutation({
    mutationFn: async (data: CreateServerRequest): Promise<Server> => {
      const response = await api.post('/servers', data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'summary'] });
      toast.success('Server created successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to create server');
    },
  });

  const updateServer = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: UpdateServerRequest }): Promise<Server> => {
      const response = await api.patch(`/servers/${id}`, data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      toast.success('Server updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to update server');
    },
  });

  const deleteServer = useMutation({
    mutationFn: async (id: number): Promise<void> => {
      await api.delete(`/servers/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'summary'] });
      toast.success('Server deleted successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete server');
    },
  });

  return {
    servers,
    summary,
    isLoading,
    error,
    createServer,
    updateServer,
    deleteServer,
  };
};
