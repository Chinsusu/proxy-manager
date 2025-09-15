import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Users, Activity, ChevronDown, ChevronRight, Server, Clock, Eye, EyeOff } from 'lucide-react';
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

interface ServerWithProxies {
  id: string;
  name: string;
  host: string;
  port: number;
  status: string;
  proxies: Proxy[];
}

export const Proxies: React.FC = () => {
  const [expandedServers, setExpandedServers] = useState<Set<string>>(new Set());
  const [visiblePasswords, setVisiblePasswords] = useState<Set<string>>(new Set());

  const { data: servers, isLoading, error } = useQuery<ServerWithProxies[]>({
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

  const togglePasswordVisibility = (proxyId: string) => {
    const newVisible = new Set(visiblePasswords);
    if (newVisible.has(proxyId)) {
      newVisible.delete(proxyId);
    } else {
      newVisible.add(proxyId);
    }
    setVisiblePasswords(newVisible);
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

  // Filter servers that have proxies
  const serversWithProxies = servers?.filter(server => server.proxies && server.proxies.length > 0) || [];
  const totalProxies = serversWithProxies.reduce((total, server) => total + server.proxies.length, 0);

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
        <div className="text-red-600 mb-4">Error loading proxies</div>
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
        <h1 className="text-2xl font-bold text-gray-900">Proxies</h1>
        <p className="text-gray-600">
          Manage your proxy connections grouped by server
          {totalProxies > 0 && (
            <span className="ml-2 text-sm bg-blue-100 text-blue-800 px-2 py-1 rounded-full">
              {totalProxies} total proxies
            </span>
          )}
        </p>
      </div>

      {serversWithProxies.length === 0 ? (
        <div className="text-center py-12">
          <Users className="mx-auto h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No proxies found</h3>
          <p className="text-gray-500">Get started by adding your first proxy server.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {serversWithProxies.map((server) => (
            <div key={server.id} className="bg-white shadow rounded-lg overflow-hidden">
              <div className="p-6">
                {/* Server Header */}
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-4">
                    <div className="flex-shrink-0">
                      <Server className="h-6 w-6 text-blue-500" />
                    </div>
                    <div>
                      <h3 className="text-lg font-medium text-gray-900">{server.name}</h3>
                      <p className="text-sm text-gray-500">
                        {server.host}:{server.port}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-4">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(server.status)}`}>
                      <Activity className="w-3 h-3 mr-1" />
                      {server.status}
                    </span>
                  </div>
                </div>

                {/* Proxies Toggle Button */}
                <button
                  onClick={() => toggleServerExpansion(server.id)}
                  className="flex items-center space-x-2 text-sm font-medium text-gray-700 hover:text-gray-900 transition-colors mb-3"
                >
                  {expandedServers.has(server.id) ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  <Users className="h-4 w-4" />
                  <span>{server.proxies.length} Proxies</span>
                </button>

                {/* Proxies List */}
                {expandedServers.has(server.id) && (
                  <div className="border-t pt-4">
                    <div className="space-y-3">
                      {server.proxies.map((proxy) => (
                        <div
                          key={proxy.id}
                          className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                        >
                          <div className="flex items-center space-x-4">
                            <div className="flex-shrink-0">
                              <div className={`w-3 h-3 rounded-full ${
                                proxy.status?.toLowerCase() === 'active' 
                                  ? 'bg-green-500' 
                                  : 'bg-red-500'
                              }`}></div>
                            </div>
                            <div className="flex-1">
                              <div className="flex items-center space-x-4">
                                <div>
                                  <div className="text-sm font-medium text-gray-900">
                                    {proxy.host}:{proxy.port}
                                  </div>
                                  <div className="text-xs text-gray-500 mt-1">
                                    ID: {proxy.id}
                                  </div>
                                </div>
                                <div className="min-w-0">
                                  <div className="text-sm text-gray-700">
                                    <span className="font-medium">Username:</span> {proxy.username}
                                  </div>
                                  <div className="text-sm text-gray-700 flex items-center space-x-2">
                                    <span className="font-medium">Password:</span>
                                    <span className="font-mono">
                                      {visiblePasswords.has(proxy.id) ? proxy.password : '••••••••'}
                                    </span>
                                    <button
                                      onClick={() => togglePasswordVisibility(proxy.id)}
                                      className="text-gray-400 hover:text-gray-600 transition-colors"
                                      title={visiblePasswords.has(proxy.id) ? 'Hide password' : 'Show password'}
                                    >
                                      {visiblePasswords.has(proxy.id) ? (
                                        <EyeOff className="h-4 w-4" />
                                      ) : (
                                        <Eye className="h-4 w-4" />
                                      )}
                                    </button>
                                  </div>
                                </div>
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center space-x-4">
                            <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(proxy.status)}`}>
                              {proxy.status}
                            </span>
                            <div className="text-xs text-gray-500 text-right">
                              <div className="flex items-center">
                                <Clock className="inline w-3 h-3 mr-1" />
                                {format(new Date(proxy.updated_at), 'MMM dd, yyyy')}
                              </div>
                              <div className="mt-1">
                                {format(new Date(proxy.updated_at), 'HH:mm')}
                              </div>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
