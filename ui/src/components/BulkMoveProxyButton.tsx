import React, { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowRight, X, Users } from 'lucide-react';
import toast from 'react-hot-toast';

interface BulkMoveProxyButtonProps {
  selectedProxyIds: string[];
  availableGroups: Array<{ id: string; name: string; color?: string; }>;
  onClearSelection: () => void;
  className?: string;
}

export const BulkMoveProxyButton: React.FC<BulkMoveProxyButtonProps> = ({
  selectedProxyIds,
  availableGroups,
  onClearSelection,
  className = "inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-purple-600 hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500 transition-colors"
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState('');

  const queryClient = useQueryClient();

  const bulkMoveProxyMutation = useMutation({
    mutationFn: async ({ proxyIds, groupId }: { proxyIds: string[]; groupId: string }) => {
      // Simulate API delay for bulk operation
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Simulate API calls for each proxy
      const movePromises = proxyIds.map(proxyId => ({
        proxyId,
        groupId,
        success: true
      }));
      
      return movePromises;
    },
    onSuccess: (results) => {
      toast.success(`Successfully moved ${results.length} proxy(ies) to group`);
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      onClearSelection();
      setIsModalOpen(false);
    },
    onError: (error: any) => {
      toast.error('Failed to move some proxies');
    },
  });

  const handleBulkMove = () => {
    if (selectedProxyIds.length === 0) return;
    if (!selectedGroupId) {
      toast.error('Please select a target group');
      return;
    }

    bulkMoveProxyMutation.mutate({ 
      proxyIds: selectedProxyIds, 
      groupId: selectedGroupId 
    });
  };

  const getSelectedGroupName = () => {
    const group = availableGroups.find(g => g.id === selectedGroupId);
    return group?.name || 'No Group';
  };

  if (selectedProxyIds.length === 0) {
    return null;
  }

  return (
    <>
      <button
        onClick={() => setIsModalOpen(true)}
        className={className}
      >
        <ArrowRight className="h-4 w-4 mr-2" />
        {bulkMoveProxyMutation.isPending 
          ? `Moving ${selectedProxyIds.length}...` 
          : `Move ${selectedProxyIds.length} to Group`
        }
      </button>

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-gray-900">
                Bulk Move Proxies to Group
              </h3>
              <button onClick={() => setIsModalOpen(false)} className="text-gray-400 hover:text-gray-600">
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Selected Proxies
                </label>
                <div className="px-3 py-2 bg-gray-50 border border-gray-200 rounded-md text-sm text-gray-900">
                  <div className="flex items-center">
                    <Users className="h-4 w-4 mr-2 text-gray-500" />
                    {selectedProxyIds.length} proxy(ies) selected
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Target Group
                </label>
                <select
                  value={selectedGroupId}
                  onChange={(e) => setSelectedGroupId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                  <option value="">Select a group...</option>
                  {availableGroups.map(group => (
                    <option key={group.id} value={group.id}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>

              {selectedGroupId && (
                <div className="bg-purple-50 border border-purple-200 rounded-md p-3">
                  <p className="text-sm text-purple-800">
                    <strong>Moving:</strong> {selectedProxyIds.length} proxy(ies) to "{getSelectedGroupName()}"
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
                onClick={handleBulkMove}
                disabled={bulkMoveProxyMutation.isPending || !selectedGroupId}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-purple-600 hover:bg-purple-700 disabled:opacity-50 transition-colors"
              >
                {bulkMoveProxyMutation.isPending ? 'Moving...' : `Move ${selectedProxyIds.length} Proxies`}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
