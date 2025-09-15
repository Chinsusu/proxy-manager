import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Mapping } from '../types';

export const Mappings: React.FC = () => {
  const { data: mappings, isLoading, error } = useQuery({
    queryKey: ['mappings'],
    queryFn: async (): Promise<Mapping[]> => {
      const response = await api.get('/mappings');
      return response.data;
    },
  });

  if (isLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600">Error loading mappings</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Mappings</h1>
        <p className="text-gray-600">Traffic routing rules and configurations</p>
      </div>

      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <ul className="divide-y divide-gray-200">
          {mappings?.map((mapping) => (
            <li key={mapping.id} className="px-6 py-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center">
                    <h3 className="text-lg font-medium text-gray-900">
                      Mapping #{mapping.id}
                    </h3>
                    <span className={`ml-3 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      mapping.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                    }`}>
                      {mapping.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>

                  <div className="mt-2 grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="bg-blue-50 px-3 py-2 rounded">
                      <div className="text-xs font-medium text-blue-700 uppercase tracking-wide">Client CIDR</div>
                      <div className="text-sm font-mono text-blue-900">{mapping.client_cidr}</div>
                    </div>
                    
                    <div className="bg-purple-50 px-3 py-2 rounded">
                      <div className="text-xs font-medium text-purple-700 uppercase tracking-wide">Destination Ports</div>
                      <div className="text-sm font-mono text-purple-900">
                        {JSON.parse(mapping.dst_ports).join(', ')}
                      </div>
                    </div>
                    
                    <div className="bg-green-50 px-3 py-2 rounded">
                      <div className="text-xs font-medium text-green-700 uppercase tracking-wide">Upstream Proxy</div>
                      <div className="text-sm text-green-900">
                        {mapping.upstream_proxy?.label || `Proxy #${mapping.upstream_proxy_id}`}
                      </div>
                      {mapping.upstream_proxy && (
                        <div className="text-xs text-green-700 font-mono">
                          {mapping.upstream_proxy.host}:{mapping.upstream_proxy.port}
                        </div>
                      )}
                    </div>
                  </div>

                  {mapping.notes && (
                    <div className="mt-3 p-3 bg-gray-50 rounded">
                      <div className="text-xs font-medium text-gray-700 uppercase tracking-wide mb-1">Notes</div>
                      <p className="text-sm text-gray-900">{mapping.notes}</p>
                    </div>
                  )}

                  {mapping.server && (
                    <div className="mt-3 pt-3 border-t border-gray-200">
                      <p className="text-sm text-gray-600">
                        Server: <span className="font-medium">{mapping.server.name}</span>
                        <span className="ml-4 text-gray-400">
                          Config v{mapping.server.config_version}
                        </span>
                      </p>
                    </div>
                  )}
                </div>

                <div className="ml-6 flex flex-col items-end space-y-2">
                  <div className="text-sm text-gray-500">
                    Created: {new Date(mapping.created_at).toLocaleDateString()}
                  </div>
                  <div className="text-sm text-gray-500">
                    Updated: {new Date(mapping.updated_at).toLocaleDateString()}
                  </div>
                </div>
              </div>

              <div className="mt-4 flex items-center text-sm text-gray-600">
                <div className="flex items-center space-x-2">
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs font-medium">
                    {mapping.client_cidr}
                  </span>
                  <span>→</span>
                  <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded text-xs font-medium">
                    Ports: {JSON.parse(mapping.dst_ports).join(', ')}
                  </span>
                  <span>→</span>
                  <span className="px-2 py-1 bg-green-100 text-green-800 rounded text-xs font-medium">
                    {mapping.upstream_proxy?.label || `Proxy #${mapping.upstream_proxy_id}`}
                  </span>
                </div>
              </div>
            </li>
          ))}
        </ul>
      </div>

      {(!mappings || mappings.length === 0) && (
        <div className="text-center py-12">
          <p className="text-gray-500">No mappings configured yet</p>
        </div>
      )}
    </div>
  );
};
