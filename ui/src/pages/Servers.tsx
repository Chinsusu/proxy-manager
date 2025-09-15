import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Server, Activity, ChevronDown, ChevronRight, Users, Clock, Plus, Edit, Trash2, Tag } from 'lucide-react';
import { format } from 'date-fns';
import { ServerModal } from '../components/ServerModal';
import { DeleteConfirmModal } from '../components/DeleteConfirmModal';
import toast from 'react-hot-toast';

interface Proxy {
  id: string;
  server_id: string;
  label: string;
  type: string;
  host: string;
  port: number;
  username: string;
  password: string;
  health: string;
  created_at: string;
  updated_at: string;
}

interface ServerData {
  id: string;
  name: string;
  tags: string[];
  wan_iface: string;
  lan_iface: string;
  last_seen_at?: string;
  status: string;
  proxies: Proxy[];
  created_at: string;
  updated_at: string;
}

export const Servers: React.FC = () => {
  const [expandedServers, setExpandedServers] = useState<Set<string>>(new Set());
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [selectedServer, setSelectedServer] = useState<ServerData | null>(null);

  const queryClient = useQueryClient();

  const { data: servers, isLoading, error } = useQuery<ServerData[]>({
    queryKey: ['servers'],
    queryFn: () => api.get('/servers').then(res => res.data),
  });

  // Create server mutation
  const createServerMutation = useMutation({
    mutationFn: (serverData: Omit<ServerData, 'id' | 'proxies' | 'created_at' | 'updated_at'>) =>
      api.post('/servers', serverData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setIsAddModalOpen(false);
      toast.success('Server created successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to create server');
    },
  });

  // Update server mutation
  const updateServerMutation = useMutation({
    mutationFn: ({ id, ...serverData }: { id: string } & Partial<ServerData>) =>
      api.patch(`/servers/${id}`, serverData),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setIsEditModalOpen(false);
      setSelectedServer(null);
      toast.success('Server updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to update server');
    },
  });

  // Delete server mutation
  const deleteServerMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/servers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setIsDeleteModalOpen(false);
      setSelectedServer(null);
      toast.success('Server deleted successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete server');
    },
  });

  const toggleServerExpansion = (serverId: string) => {
    const newExpanded = new Set(expandedServers);
    if (newExpanded.has(serverId)) {
      newExpanded.delete(serverId);
    } else {
      newExpanded.add(serverId);
    }
    setExpandedServers(newExpanded);
  };

  const handleAddServer = () => {
    setSelectedServer(null);
    setIsAddModalOpen(true);
  };

  const handleEditServer = (server: ServerData) => {
    // Parse tags from JSON string if needed
    let tags = [];
    if (typeof server.tags === 'string') {
      try {
        tags = JSON.parse(server.tags);
      } catch {
        tags = [];
      }
    } else if (Array.isArray(server.tags)) {
      tags = server.tags;
    }
    
    setSelectedServer({
      ...server,
      tags
    });
    setIsEditModalOpen(true);
  };

  const handleDeleteServer = (server: ServerData) => {
    if (server.id === '0') return; // Cannot delete unassigned server
    setSelectedServer(server);
    setIsDeleteModalOpen(true);
  };

  const handleServerSubmit = (serverData: any) => {
    if (selectedServer && isEditModalOpen) {
      updateServerMutation.mutate({ ...serverData, id: selectedServer.id });
    } else {
      createServerMutation.mutate(serverData);
    }
  };

  const handleConfirmDelete = () => {
    if (selectedServer) {
      deleteServerMutation.mutate(selectedServer.id);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'online':
        return 'bg-green-100 text-green-800';
      case 'offline':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const parseServerTags = (tags: any): string[] => {
    if (typeof tags === 'string') {
      try {
        return JSON.parse(tags);
      } catch {
        return [];
      }
    }
    return Array.isArray(tags) ? tags : [];
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 mb-4">Error loading servers</div>
        <button
          onClick={() => window.location.reload()}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Servers</h1>
            <p className="text-gray-600">Manage your proxy servers</p>
          </div>
          <button
            onClick={handleAddServer}
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Server
          </button>
        </div>
      </div>

      {!servers || servers.length === 0 ? (
        <div className="text-center py-12">
          <Server className="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No servers found</h3>
          <p className="text-gray-500 mb-4">Get started by adding your first server.</p>
          <button
            onClick={handleAddServer}
            className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Server
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          console.log("All servers:", servers);
          {servers.filter(server => true).map((server) => {
          console.log("Filtered servers:", servers.filter(server => true));
            const serverTags = parseServerTags(server.tags);
            
            return (
              <div key={server.id} className="bg-white shadow rounded-lg overflow-hidden">
                <div className="p-6">
                  {/* Server Header */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <div className="flex-shrink-0">
                        <Server className="h-6 w-6 text-blue-500" />
                      </div>
                      <div>
                        <div className="flex items-center space-x-3">
                          <h3 className="text-lg font-medium text-gray-900">{server.name === 'Unassigned' ? '📦 Imported Proxies' : server.name}</h3>
                          <div className="flex items-center space-x-1">
                            <button
                              onClick={() => handleEditServer(server)}
                              className="p-1 text-gray-400 hover:text-blue-600 transition-colors"
                              title="Edit server"
                            >
                              <Edit className="h-4 w-4" />
                            </button>
                            <button
                              disabled={server.id === '0'}
                              onClick={() => handleDeleteServer(server)}
                              className={`p-1 transition-colors ${
                                server.id === '0' 
                                  ? "text-gray-300 cursor-not-allowed" 
                                  : "text-gray-400 hover:text-red-600"
                              }`}
                              title={server.id === '0' ? "Cannot delete unassigned server" : "Delete server"}
                            >
                              <Trash2 className="h-4 w-4" />
                            </button>
                          </div>
                        </div>
                        <div className="flex items-center space-x-4 mt-1">
                          <p className="text-sm text-gray-500">
                            {server.wan_iface && `WAN: ${server.wan_iface}`}
                            {server.wan_iface && server.lan_iface && ' • '}
                            {server.lan_iface && `LAN: ${server.lan_iface}`}
                          </p>
                        </div>
                        {serverTags.length > 0 && (
                          <div className="flex items-center space-x-1 mt-2">
                            <Tag className="h-3 w-3 text-gray-400" />
                            <div className="flex flex-wrap gap-1">
                              {serverTags.map((tag, index) => (
                                <span
                                  key={index}
                                  className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800"
                                >
                                  {tag}
                                </span>
                              ))}
                            </div>
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center space-x-4">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(server.status)}`}>
                        <Activity className="w-3 h-3 mr-1" />
                        {server.status}
                      </span>
                      <div className="text-sm text-gray-500">
                        <Clock className="inline w-4 h-4 mr-1" />
                        {server.last_seen_at 
                          ? format(new Date(server.last_seen_at), 'MMM dd, yyyy')
                          : format(new Date(server.updated_at), 'MMM dd, yyyy')
                        }
                      </div>
                    </div>
                  </div>

                  {/* Proxies Section */}
                  {server.proxies && server.proxies.length > 0 && (
                    <div className="mt-4">
                      <button
                        onClick={() => toggleServerExpansion(server.id)}
                        className="flex items-center space-x-2 text-sm font-medium text-gray-700 hover:text-gray-900 transition-colors"
                      >
                        {expandedServers.has(server.id) ? (
                          <ChevronDown className="h-4 w-4" />
                        ) : (
                          <ChevronRight className="h-4 w-4" />
                        )}
                        <Users className="h-4 w-4" />
                        <span>{server.proxies.length} Proxies</span>
                      </button>

                      {expandedServers.has(server.id) && (
                        <div className="mt-3 border-t pt-3">
                          <div className="grid gap-3">
                            {server.proxies.map((proxy) => (
                              <div
                                key={proxy.id}
                                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                              >
                                <div className="flex items-center space-x-3">
                                  <div className="flex-shrink-0">
                                    <div className={`w-2 h-2 rounded-full ${
                                      proxy.health === 'ok' ? 'bg-green-500' : 'bg-red-500'
                                    }`}></div>
                                  </div>
                                  <div>
                                    <div className="text-sm font-medium text-gray-900">
                                      {proxy.label} ({proxy.type})
                                    </div>
                                    <div className="text-xs text-gray-500">
                                      {proxy.host}:{proxy.port}
                                      {proxy.username && ` • ${proxy.username}`}
                                    </div>
                                  </div>
                                </div>
                                <div className="flex items-center space-x-3">
                                  <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                    proxy.health === 'ok' 
                                      ? 'bg-green-100 text-green-800'
                                      : proxy.health === 'fail'
                                      ? 'bg-red-100 text-red-800'
                                      : 'bg-gray-100 text-gray-800'
                                  }`}>
                                    {proxy.health}
                                  </span>
                                  <div className="text-xs text-gray-500">
                                    {format(new Date(proxy.updated_at), 'MMM dd')}
                                  </div>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Server Info Footer */}
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <div className="flex justify-between text-sm text-gray-500">
                      <div>Created: {format(new Date(server.created_at), 'MMM dd, yyyy')}</div>
                      <div>Updated: {format(new Date(server.updated_at), 'MMM dd, yyyy')}</div>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Modals */}
      <ServerModal
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        onSubmit={handleServerSubmit}
        isLoading={createServerMutation.isPending}
      />

      <ServerModal
        isOpen={isEditModalOpen}
        onClose={() => {
          setIsEditModalOpen(false);
          setSelectedServer(null);
        }}
        onSubmit={handleServerSubmit}
        server={selectedServer}
        isLoading={updateServerMutation.isPending}
      />

      <DeleteConfirmModal
        isOpen={isDeleteModalOpen}
        onClose={() => {
          setIsDeleteModalOpen(false);
          setSelectedServer(null);
        }}
        onConfirm={handleConfirmDelete}
        title="Delete Server"
        message={`Are you sure you want to delete "${selectedServer?.name}"? ${
          selectedServer?.proxies?.length 
            ? `This will also affect ${selectedServer.proxies.length} associated proxies.` 
            : 'This action cannot be undone.'
        }`}
        isLoading={deleteServerMutation.isPending}
      />
    </div>
  );
};
