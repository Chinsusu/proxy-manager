import React from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2 } from 'lucide-react';
import { api } from '../lib/api';
import toast from 'react-hot-toast';

interface BulkDeleteProxyButtonProps {
  selectedProxyIds: string[];
  onClearSelection: () => void;
  className?: string;
}

export const BulkDeleteProxyButton: React.FC<BulkDeleteProxyButtonProps> = ({
  selectedProxyIds,
  onClearSelection,
  className = "inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 transition-colors"
}) => {
  const queryClient = useQueryClient();

  const bulkDeleteMutation = useMutation({
    mutationFn: async (proxyIds: string[]) => {
      // Delete proxies sequentially to avoid overwhelming the API
      const deletePromises = proxyIds.map(proxyId => 
        api.delete(`/proxies/${proxyId}`)
      );
      return Promise.all(deletePromises);
    },
    onSuccess: (results) => {
      toast.success(`Successfully deleted ${results.length} proxy(ies)`);
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      onClearSelection();
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete some proxies');
    },
  });

  const handleBulkDelete = () => {
    if (selectedProxyIds.length === 0) return;

    const confirmMessage = `Are you sure you want to delete ${selectedProxyIds.length} proxy(ies)? This action cannot be undone.`;
    
    if (window.confirm(confirmMessage)) {
      bulkDeleteMutation.mutate(selectedProxyIds);
    }
  };

  if (selectedProxyIds.length === 0) {
    return null;
  }

  return (
    <button
      onClick={handleBulkDelete}
      disabled={bulkDeleteMutation.isPending}
      className={className}
    >
      <Trash2 className="h-4 w-4 mr-2" />
      {bulkDeleteMutation.isPending 
        ? `Deleting ${selectedProxyIds.length}...` 
        : `Delete ${selectedProxyIds.length} Selected`
      }
    </button>
  );
};
