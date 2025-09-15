import React, { useState, useEffect } from 'react';
import { X, Server } from 'lucide-react';

interface ServerData {
  id?: string;
  name: string;
  tags: string[];
  wan_iface: string;
  lan_iface: string;
  status: string;
}

interface ServerModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: ServerData) => void;
  server?: ServerData | null;
  isLoading?: boolean;
}

export const ServerModal: React.FC<ServerModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  server,
  isLoading = false
}) => {
  const [formData, setFormData] = useState<ServerData>({
    name: '',
    tags: [],
    wan_iface: '',
    lan_iface: '',
    status: 'offline'
  });

  const [tagsInput, setTagsInput] = useState<string>('');

  useEffect(() => {
    if (server) {
      setFormData(server);
      setTagsInput(server.tags.join(', '));
    } else {
      setFormData({
        name: '',
        tags: [],
        wan_iface: '',
        lan_iface: '',
        status: 'offline'
      });
      setTagsInput('');
    }
  }, [server]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // Parse tags from comma-separated input
    const tags = tagsInput
      .split(',')
      .map(tag => tag.trim())
      .filter(tag => tag.length > 0);
    
    onSubmit({
      ...formData,
      tags
    });
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleTagsChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTagsInput(e.target.value);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <Server className="h-5 w-5 text-blue-500" />
            <h3 className="text-lg font-medium text-gray-900">
              {server ? 'Edit Server' : 'Add New Server'}
            </h3>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Server Name *
              </label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter server name"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tags
              </label>
              <input
                type="text"
                value={tagsInput}
                onChange={handleTagsChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter tags separated by commas (e.g., production, us-east)"
              />
              <p className="text-xs text-gray-500 mt-1">Separate multiple tags with commas</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                WAN Interface
              </label>
              <input
                type="text"
                name="wan_iface"
                value={formData.wan_iface}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="e.g., eth0, wlan0"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                LAN Interface
              </label>
              <input
                type="text"
                name="lan_iface"
                value={formData.lan_iface}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="e.g., eth1, br0"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Status
              </label>
              <select
                name="status"
                value={formData.status}
                onChange={handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
              >
                <option value="offline">Offline</option>
                <option value="online">Online</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end space-x-3 mt-6">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
              disabled={isLoading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors disabled:opacity-50"
              disabled={isLoading}
            >
              {isLoading ? 'Saving...' : (server ? 'Update' : 'Create')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
