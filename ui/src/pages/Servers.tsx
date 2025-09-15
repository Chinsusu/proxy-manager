import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Server, Activity, ChevronDown, ChevronRight, Users, Clock } from 'lucide-react';
import { format } from 'date-fns';

interface Proxy {
  id: string;
  host: string;
  port: number;
  username: string;
  password: string;
  status: string;
  server_id: string;
  created_at: string;
  updated_at: string;
}

interface ServerData {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  status: string;
  proxies: Proxy[];
  created_at: string;
  updated_at: string;
}

export const Servers: React.FC = () => {
  const [expandedServers, setExpandedServers] = useState<Set<string>>(new Set());

  const { data: servers, isLoading, error } = useQuery<ServerData[]>({
    queryKey: ['servers'],
    queryFn: () => api.get('/servers').then(res => res.data),
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

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'inactive':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
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
        <h1 className="text-2xl font-bold text-gray-900">Servers</h1>
        <p className="text-gray-600">Manage your proxy servers</p>
      </div>

      {!servers || servers.length === 0 ? (
        <div className="text-center py-12">
          <Server className="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No servers found</h3>
          <p className="text-gray-500">Get started by adding your first server.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {servers.map((server) => (
            <div key={server.id} className="bg-white shadow rounded-lg overflow-hidden">
              <div className="p-6">
                {/* Server Header */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className="flex-shrink-0">
                      <Server className="h-6 w-6 text-blue-500" />
                    </div>
                    <div>
                      <h3 className="text-lg font-medium text-gray-900">{server.name}</h3>
                      <p className="text-sm text-gray-500">
                        {server.host}:{server.port} • {server.username}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(server.status)}`}>
                      <Activity className="w-3 h-3 mr-1" />
                      {server.status}
                    </span>
                    <div className="text-sm text-gray-500">
                      <Clock className="inline w-4 h-4 mr-1" />
                      {format(new Date(server.updated_at), 'MMM dd, yyyy')}
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
                                  <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                                </div>
                                <div>
                                  <div className="text-sm font-medium text-gray-900">
                                    {proxy.host}:{proxy.port}
                                  </div>
                                  <div className="text-xs text-gray-500">
                                    {proxy.username} • ****
                                  </div>
                                </div>
                              </div>
                              <div className="flex items-center space-x-3">
                                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(proxy.status)}`}>
                                  {proxy.status}
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
          ))}
        </div>
      )}
    </div>
  );
};
