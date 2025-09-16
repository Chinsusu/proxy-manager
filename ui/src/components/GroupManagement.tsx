import React, { useState } from 'react';
import { api } from '../lib/api';
import { Edit, Trash2, Plus } from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'react-hot-toast';

export interface ProxyGroup {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  proxies?: any[];
}

export function GroupManagement() {
  const [isCreating, setIsCreating] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [newGroupName, setNewGroupName] = useState('');
  const [newGroupDescription, setNewGroupDescription] = useState('');
  const [editGroupName, setEditGroupName] = useState('');
  const [editGroupDescription, setEditGroupDescription] = useState('');
  
  const queryClient = useQueryClient();

  // Fetch groups
  const { data: groups = [], isLoading } = useQuery({
    queryKey: ['groups'],
    queryFn: () => api.get('/groups').then(res => res.data)
  });

  // Create group mutation
  const createGroupMutation = useMutation({
    mutationFn: (groupData: { name: string; description: string }) =>
      api.post('/groups', groupData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setIsCreating(false);
      setNewGroupName('');
      setNewGroupDescription('');
      toast.success('Group created successfully');
    },
    onError: () => {
      toast.error('Failed to create group');
    }
  });

  // Update group mutation
  const updateGroupMutation = useMutation({
    mutationFn: ({ id, ...groupData }: { id: number; name: string; description: string }) =>
      api.put(`/groups/${id}`, groupData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      setEditingId(null);
      toast.success('Group updated successfully');
    },
    onError: () => {
      toast.error('Failed to update group');
    }
  });

  // Delete group mutation
  const deleteGroupMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/groups/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      toast.success('Group deleted successfully');
    },
    onError: () => {
      toast.error('Failed to delete group');
    }
  });

  const handleCreateGroup = () => {
    if (newGroupName.trim()) {
      createGroupMutation.mutate({
        name: newGroupName.trim(),
        description: newGroupDescription.trim()
      });
    }
  };

  const handleUpdateGroup = (id: number) => {
    if (editGroupName.trim()) {
      updateGroupMutation.mutate({
        id,
        name: editGroupName.trim(),
        description: editGroupDescription.trim()
      });
    }
  };

  const handleDeleteGroup = (id: number) => {
    if (window.confirm('Are you sure you want to delete this group?')) {
      deleteGroupMutation.mutate(id);
    }
  };

  const startEditing = (group: ProxyGroup) => {
    setEditingId(group.id);
    setEditGroupName(group.name);
    setEditGroupDescription(group.description || '');
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditGroupName('');
    setEditGroupDescription('');
  };

  if (isLoading) {
    return <div>Loading groups...</div>;
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="mb-4">
        <h3 className="text-lg font-semibold">Proxy Groups</h3>
        <p className="text-gray-600">Manage proxy groups</p>
      </div>
      
      <div className="space-y-4">
        {/* Create new group form */}
        {isCreating ? (
          <div className="p-4 border rounded-lg bg-gray-50">
            <div className="space-y-3">
              <input
                type="text"
                placeholder="Group name"
                value={newGroupName}
                onChange={(e) => setNewGroupName(e.target.value)}
                className="w-full px-3 py-2 border rounded-md"
              />
              <textarea
                placeholder="Group description (optional)"
                value={newGroupDescription}
                onChange={(e) => setNewGroupDescription(e.target.value)}
                className="w-full px-3 py-2 border rounded-md"
                rows={3}
              />
              <div className="flex gap-2">
                <button 
                  onClick={handleCreateGroup}
                  disabled={!newGroupName.trim() || createGroupMutation.isPending}
                  className="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {createGroupMutation.isPending ? 'Creating...' : 'Create'}
                </button>
                <button 
                  onClick={() => {
                    setIsCreating(false);
                    setNewGroupName('');
                    setNewGroupDescription('');
                  }}
                  className="px-3 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        ) : (
          <button
            onClick={() => setIsCreating(true)}
            className="flex items-center gap-2 px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            <Plus className="w-4 h-4" />
            New Group
          </button>
        )}

        {/* Groups list */}
        {groups.length === 0 ? (
          <p className="text-gray-500 text-center py-4">No groups found. Create your first group!</p>
        ) : (
          <div className="space-y-2">
            {groups.map((group: ProxyGroup) => (
              <div key={group.id} className="p-3 border rounded-lg">
                {editingId === group.id ? (
                  <div className="space-y-3">
                    <input
                      type="text"
                      value={editGroupName}
                      onChange={(e) => setEditGroupName(e.target.value)}
                      className="w-full px-3 py-2 border rounded-md"
                    />
                    <textarea
                      value={editGroupDescription}
                      onChange={(e) => setEditGroupDescription(e.target.value)}
                      className="w-full px-3 py-2 border rounded-md"
                      rows={3}
                    />
                    <div className="flex gap-2">
                      <button 
                        onClick={() => handleUpdateGroup(group.id)}
                        disabled={!editGroupName.trim() || updateGroupMutation.isPending}
                        className="px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                      >
                        {updateGroupMutation.isPending ? 'Saving...' : 'Save'}
                      </button>
                      <button 
                        onClick={cancelEditing}
                        className="px-3 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium">{group.name}</h4>
                      {group.description && (
                        <p className="text-sm text-gray-600">{group.description}</p>
                      )}
                      <p className="text-xs text-gray-400">
                        {group.proxies?.length || 0} proxies
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => startEditing(group)}
                        className="p-2 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded-md"
                      >
                        <Edit className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => handleDeleteGroup(group.id)}
                        disabled={deleteGroupMutation.isPending}
                        className="p-2 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded-md disabled:opacity-50"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
