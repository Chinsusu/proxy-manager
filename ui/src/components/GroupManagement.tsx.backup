import React, { useState } from 'react';
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { Plus, Edit2, Trash2, Users, Palette, X } from 'lucide-react';
import { api } from '../lib/api';
import toast from 'react-hot-toast';

interface ProxyGroup {
  id: string;
  name: string;
  description?: string;
  color?: string;
  created_at: string;
  updated_at: string;
}

interface GroupManagementProps {
  onGroupCreate?: (group: ProxyGroup) => void;
  onGroupUpdate?: (group: ProxyGroup) => void;
  onGroupDelete?: (groupId: string) => void;
}

const GROUP_COLORS = [
  '#3B82F6', // blue
  '#10B981', // green  
  '#F59E0B', // yellow
  '#EF4444', // red
  '#8B5CF6', // purple
  '#F97316', // orange
  '#06B6D4', // cyan
  '#84CC16', // lime
];

export const GroupManagement: React.FC<GroupManagementProps> = ({
  onGroupCreate,
  onGroupUpdate, 
  onGroupDelete
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState<ProxyGroup | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    color: GROUP_COLORS[0]
  });

  const queryClient = useQueryClient();

  // Fetch groups (mock API for now)
  const { data: groups = [], isLoading } = useQuery<ProxyGroup[]>({
    queryKey: ['groups'],
    queryFn: async () => {
      // Mock data - in real implementation this would be API call
      return [
        {
          id: '1',
          name: 'Default',
          description: 'Default proxy group',
          color: '#3B82F6',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      ];
    }
  });

  const createGroupMutation = useMutation({
    mutationFn: async (groupData: Omit<ProxyGroup, 'id' | 'created_at' | 'updated_at'>) => {
      // Mock API call - replace with real API
      const newGroup: ProxyGroup = {
        id: Date.now().toString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        ...groupData
      };
      return newGroup;
    },
    onSuccess: (newGroup) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      toast.success('Group created successfully');
      onGroupCreate?.(newGroup);
      closeModal();
    },
    onError: () => {
      toast.error('Failed to create group');
    }
  });

  const updateGroupMutation = useMutation({
    mutationFn: async (groupData: ProxyGroup) => {
      // Mock API call - replace with real API
      return { ...groupData, updated_at: new Date().toISOString() };
    },
    onSuccess: (updatedGroup) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      toast.success('Group updated successfully');
      onGroupUpdate?.(updatedGroup);
      closeModal();
    },
    onError: () => {
      toast.error('Failed to update group');
    }
  });

  const deleteGroupMutation = useMutation({
    mutationFn: async (groupId: string) => {
      // Mock API call - replace with real API
      return groupId;
    },
    onSuccess: (groupId) => {
      queryClient.invalidateQueries({ queryKey: ['groups'] });
      toast.success('Group deleted successfully');
      onGroupDelete?.(groupId);
    },
    onError: () => {
      toast.error('Failed to delete group');
    }
  });

  const openModal = (group?: ProxyGroup) => {
    if (group) {
      setEditingGroup(group);
      setFormData({
        name: group.name,
        description: group.description || '',
        color: group.color || GROUP_COLORS[0]
      });
    } else {
      setEditingGroup(null);
      setFormData({
        name: '',
        description: '',
        color: GROUP_COLORS[0]
      });
    }
    setIsModalOpen(true);
  };

  const closeModal = () => {
    setIsModalOpen(false);
    setEditingGroup(null);
    setFormData({ name: '', description: '', color: GROUP_COLORS[0] });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      toast.error('Group name is required');
      return;
    }

    if (editingGroup) {
      updateGroupMutation.mutate({
        ...editingGroup,
        ...formData
      });
    } else {
      createGroupMutation.mutate(formData);
    }
  };

  const handleDelete = (group: ProxyGroup) => {
    if (window.confirm(`Are you sure you want to delete group "${group.name}"? This will ungroup all proxies in this group.`)) {
      deleteGroupMutation.mutate(group.id);
    }
  };

  if (isLoading) {
    return <div>Loading groups...</div>;
  }

  return (
    <div className="mb-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center space-x-2">
          <Users className="h-5 w-5 text-gray-600" />
          <h2 className="text-lg font-medium text-gray-900">Proxy Groups</h2>
        </div>
        <button
          onClick={() => openModal()}
          className="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-blue-600 hover:bg-blue-700 transition-colors"
        >
          <Plus className="h-3 w-3 mr-1" />
          New Group
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {groups.map((group) => (
          <div key={group.id} className="bg-white border rounded-lg p-4 hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center space-x-2">
                <div 
                  className="w-3 h-3 rounded-full" 
                  style={{ backgroundColor: group.color || '#3B82F6' }}
                />
                <h3 className="font-medium text-gray-900">{group.name}</h3>
              </div>
              <div className="flex items-center space-x-1">
                <button
                  onClick={() => openModal(group)}
                  className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
                  title="Edit group"
                >
                  <Edit2 className="h-3 w-3" />
                </button>
                {group.name !== 'Default' && (
                  <button
                    onClick={() => handleDelete(group)}
                    className="p-1 text-gray-400 hover:text-red-600 transition-colors"
                    title="Delete group"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                )}
              </div>
            </div>
            {group.description && (
              <p className="text-xs text-gray-500 mb-2">{group.description}</p>
            )}
            <div className="text-xs text-gray-400">
              Created: {new Date(group.created_at).toLocaleDateString()}
            </div>
          </div>
        ))}
      </div>

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-gray-900">
                {editingGroup ? 'Edit Group' : 'Create New Group'}
              </h3>
              <button onClick={closeModal} className="text-gray-400 hover:text-gray-600">
                <X className="h-5 w-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Group Name
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Enter group name"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description (optional)
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Enter group description"
                  rows={3}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <Palette className="inline h-4 w-4 mr-1" />
                  Color
                </label>
                <div className="flex space-x-2">
                  {GROUP_COLORS.map((color) => (
                    <button
                      key={color}
                      type="button"
                      onClick={() => setFormData({ ...formData, color })}
                      className={`w-8 h-8 rounded-full border-2 ${
                        formData.color === color ? 'border-gray-400' : 'border-transparent'
                      }`}
                      style={{ backgroundColor: color }}
                    />
                  ))}
                </div>
              </div>

              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={closeModal}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={createGroupMutation.isPending || updateGroupMutation.isPending}
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 disabled:opacity-50 transition-colors"
                >
                  {editingGroup ? 'Update' : 'Create'} Group
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
