import React from 'react';
import { ArrowRight } from 'lucide-react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { toast } from 'react-hot-toast';
import { api } from '../lib/api';

export interface Proxy {
  id: string | number;
  label: string | undefined;
  type: string | undefined;
  host: string | undefined;
  port: number | undefined;
  group_id?: number;
}

interface MoveProxyButtonProps {
  proxy: any;
  availableGroups?: { id: string | number; name: string }[];
}

export function MoveProxyButton({ proxy, availableGroups = [] }: MoveProxyButtonProps) {
  const [selectedGroupId, setSelectedGroupId] = React.useState<string>('');
  const queryClient = useQueryClient();

  // Fetch groups if not provided
  const { data: groups = [] } = useQuery({
    queryKey: ['groups'],
    queryFn: () => api.get('/groups').then(res => res.data),
    enabled: availableGroups.length === 0
  });

  const finalGroups = availableGroups.length > 0 ? availableGroups : groups;

  const moveProxyMutation = useMutation({
    mutationFn: ({ proxyId, groupId }: { proxyId: number; groupId: number | null }) =>
      api.put(`/proxies/${proxyId}/group`, { group_id: groupId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['proxies'] });
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setSelectedGroupId('');
      toast.success(`Proxy "${proxy.label}" moved successfully`);
    },
    onError: (error: any) => {
      console.error('Move error:', error);
      toast.error(`Failed to move proxy: ${error.response?.data?.error || error.message}`);
    }
  });

  const handleMoveProxy = () => {
    if (selectedGroupId === '') return;
    
    const groupId = selectedGroupId === 'null' ? null : parseInt(selectedGroupId);
    moveProxyMutation.mutate({
      proxyId: typeof proxy.id === "string" ? parseInt(proxy.id) : proxy.id,
      groupId: groupId
    });
  };

  return (
    <div className="flex items-center gap-2">
      <select
        value={selectedGroupId}
        onChange={(e) => setSelectedGroupId(e.target.value)}
        className="px-3 py-2 border border-gray-300 rounded-md text-sm w-48"
      >
        <option value="">Select group</option>
        <option value="null">No Group</option>
        {finalGroups.map((group: any) => (
          <option key={group.id} value={group.id.toString()}>
            {group.name}
          </option>
        ))}
      </select>
      
      <button
        onClick={handleMoveProxy}
        disabled={selectedGroupId === '' || moveProxyMutation.isPending}
        className="flex items-center gap-1 px-3 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <ArrowRight className="w-4 h-4" />
        {moveProxyMutation.isPending ? 'Moving...' : 'Move'}
      </button>
    </div>
  );
}
