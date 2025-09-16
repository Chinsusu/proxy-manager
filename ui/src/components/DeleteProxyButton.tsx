import React from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2 } from 'lucide-react';
import { api } from '../lib/api';
import toast from 'react-hot-toast';

interface DeleteProxyButtonProps {
  proxyId: string;
  proxyLabel: string;
  className?: string;
}

export const DeleteProxyButton: React.FC<DeleteProxyButtonProps> = ({
  proxyId,
  proxyLabel,
  className = "p-1 text-gray-400 hover:text-red-600 transition-colors"
}) => {
  const queryClient = useQueryClient();

  const deleteProxyMutation = useMutation({
    mutationFn: (proxyId: string) => api.delete(`/proxies/${proxyId}`),
    onSuccess: () => {
      toast.success('Proxy deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['servers'] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete proxy');
    },
  });

  const handleDelete = () => {
    if (window.confirm(`Are you sure you want to delete proxy "${proxyLabel}"? This action cannot be undone.`)) {
      deleteProxyMutation.mutate(proxyId);
    }
  };

  return (
    <button
      onClick={handleDelete}
      disabled={deleteProxyMutation.isPending}
      className={className}
      title="Delete proxy"
    >
      <Trash2 className="h-4 w-4" />
    </button>
  );
};
