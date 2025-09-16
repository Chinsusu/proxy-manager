import React, { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowRight, X } from 'lucide-react';
import { api } from '../lib/api';
import toast from 'react-hot-toast';

interface MoveProxyButtonProps {
  proxyId: string;
  proxyLabel: string;
  currentGroupId?: string;
  availableGroups: Array<{ id: string; name: string; color?: string; }>;
  className?: string;
}

export const MoveProxyButton: React.FC<MoveProxyButtonProps> = ({
  proxyId,
  proxyLabel,
  currentGroupId,
  availableGroups,
  className = "p-1 text-gray-400 hover:text-blue-600 transition-colors"
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState(currentGroupId || '');

  const queryClient = useQueryClient();

  const moveProxyMutation = useMutation({
    mutationFn: async ({ proxyId, groupId }: { proxyId: string; groupId: string }) => {
      // Mock API call - replace with real API
      return api.put(`/proxies/${proxyId}`, { group_id: groupId || null });
    },
    onSuccess: () => {
      toast.success('Proxy moved successfully');
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setIsModalOpen(false);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to move proxy');
    },
  });

  const handleMove = () => {
    if (selectedGroupId === currentGroupId) {
      toast.success('Proxy is already in the selected group');
      setIsModalOpen(false);
      return;
    }

    moveProxyMutation.mutate({ proxyId, groupId: selectedGroupId });
  };

  const getCurrentGroupName = () => {
    const group = availableGroups.find(g => g.id === currentGroupId);
    return group?.name || 'No Group';
  };

  const getSelectedGroupName = () => {
    const group = availableGroups.find(g => g.id === selectedGroupId);
    return group?.name || 'No Group';
  };

  return (
    <>
      <button
        onClick={() => setIsModalOpen(true)}
        className={className}
        title="Move to group"
      >
        <ArrowRight className="h-4 w-4" />
      </button>

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-gray-900">
                Move Proxy to Group
              </h3>
              <button onClick={() => setIsModalOpen(false)} className="text-gray-400 hover:text-gray-600">
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Proxy
                </label>
                <div className="px-3 py-2 bg-gray-50 border border-gray-200 rounded-md text-sm text-gray-900">
                  {proxyLabel}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Current Group
                </label>
                <div className="px-3 py-2 bg-gray-50 border border-gray-200 rounded-md text-sm text-gray-600">
                  {getCurrentGroupName()}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Move to Group
                </label>
                <select
                  value={selectedGroupId}
                  onChange={(e) => setSelectedGroupId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="">No Group</option>
                  {availableGroups.map(group => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>

              {selectedGroupId !== currentGroupId && (
                <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
                  <p className="text-sm text-blue-800">
                    <strong>Moving:</strong> {proxyLabel} from "{getCurrentGroupName()}" to "{getSelectedGroupName()}"
                  </p>
                </div>
              )}
            </div>

            <div className="flex justify-end space-x-3 pt-6">
              <button
                onClick={() => setIsModalOpen(false)}
                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleMove}
                disabled={moveProxyMutation.isPending || selectedGroupId === currentGroupId}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                {moveProxyMutation.isPending ? 'Moving...' : 'Move Proxy'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
