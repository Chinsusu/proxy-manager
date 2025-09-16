import React, { useState } from 'react';
import { ArrowRightLeft } from 'lucide-react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { toast } from 'react-hot-toast';
import { api } from '../lib/api';

interface BulkMoveProxyButtonProps {
  selectedProxyIds: number[];
  availableGroups?: { id: number; name: string }[];
}

export function BulkMoveProxyButton({ selectedProxyIds, availableGroups = [] }: BulkMoveProxyButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedGroupId, setSelectedGroupId] = useState<string>('');
  const queryClient = useQueryClient();

  // Fetch groups if not provided
  const { data: groups = [] } = useQuery({
    queryKey: ['groups'],
    queryFn: () => api.get('/groups').then(res => res.data),
    enabled: availableGroups.length === 0
  });

  const finalGroups = availableGroups.length > 0 ? availableGroups : groups;

  const bulkMoveProxyMutation = useMutation({
    mutationFn: ({ proxyIds, groupId }: { proxyIds: number[]; groupId: number | null }) =>
      api.put('/proxies/bulk-move', { proxy_ids: proxyIds, group_id: groupId }),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['proxies'] });
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setSelectedGroupId('');
      setIsOpen(false);
      const movedCount = response.data?.moved_count || selectedProxyIds.length;
      toast.success(`Successfully moved ${movedCount} proxies`);
    },
    onError: (error: any) => {
      console.error('Bulk move error:', error);
      toast.error(`Failed to move proxies: ${error.response?.data?.error || error.message}`);
    }
  });

  const handleBulkMoveProxies = () => {
    if (selectedGroupId === '' || selectedProxyIds.length === 0) return;
    
    const groupId = selectedGroupId === 'null' ? null : parseInt(selectedGroupId);
    bulkMoveProxyMutation.mutate({
      proxyIds: selectedProxyIds,
      groupId: groupId
    });
  };

  if (selectedProxyIds.length === 0) {
    return null;
  }

  return (
    <>
      <button
        onClick={() => setIsOpen(true)}
        className="flex items-center gap-2 px-3 py-2 text-sm border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
      >
        <ArrowRightLeft className="w-4 h-4" />
        Bulk Move ({selectedProxyIds.length})
      </button>

      {/* Modal */}
      {isOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-96 max-w-90vw">
            <div className="mb-4">
              <h3 className="text-lg font-semibold">Move {selectedProxyIds.length} Proxies to Group</h3>
            </div>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-2">Select target group:</label>
                <select
                  value={selectedGroupId}
                  onChange={(e) => setSelectedGroupId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="">Select group</option>
                  <option value="null">No Group</option>
                  {finalGroups.map((group: any) => (
                    <option key={group.id} value={group.id.toString()}>
                      {group.name}
                    </option>
                  ))}
                </select>
              </div>
              
              <div className="flex justify-end gap-2">
                <button 
                  onClick={() => setIsOpen(false)}
                  className="px-3 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleBulkMoveProxies}
                  disabled={selectedGroupId === '' || bulkMoveProxyMutation.isPending}
                  className="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {bulkMoveProxyMutation.isPending ? 'Moving...' : `Move ${selectedProxyIds.length} Proxies`}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
