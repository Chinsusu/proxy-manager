import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Proxy } from '../types';

export const Proxies: React.FC = () => {
  const { data: proxies, isLoading, error } = useQuery({
    queryKey: ['proxies'],
    queryFn: async (): Promise<Proxy[]> => {
      const response = await api.get('/proxies');
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
        <p className="text-red-600">Error loading proxies</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Proxies</h1>
        <p className="text-gray-600">Manage your proxy configurations</p>
      </div>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {proxies?.map((proxy) => (
          <div key={proxy.id} className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="h-8 w-8 bg-blue-100 rounded-full flex items-center justify-center">
                    <span className="text-blue-600 font-semibold text-sm">
                      {proxy.type.charAt(0).toUpperCase()}
                    </span>
                  </div>
                </div>
                <div className="ml-4 flex-1">
                  <h3 className="text-lg font-medium text-gray-900">{proxy.label}</h3>
                  <p className="text-sm text-gray-600">{proxy.type.toUpperCase()}</p>
                </div>
                <div className="ml-4">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    proxy.health === 'ok' ? 'bg-green-100 text-green-800' : 
                    proxy.health === 'fail' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'
                  }`}>
                    {proxy.health}
                  </span>
                </div>
              </div>

              <div className="mt-4">
                <div className="bg-gray-50 px-3 py-2 rounded">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-gray-700">Host:</span>
                    <span className="text-sm text-gray-900">{proxy.host}</span>
                  </div>
                  <div className="flex items-center justify-between mt-1">
                    <span className="text-sm font-medium text-gray-700">Port:</span>
                    <span className="text-sm text-gray-900">{proxy.port}</span>
                  </div>
                  {proxy.username && (
                    <>
                      <div className="flex items-center justify-between mt-1">
                        <span className="text-sm font-medium text-gray-700">Username:</span>
                        <span className="text-sm text-gray-900">{proxy.username}</span>
                      </div>
                      <div className="flex items-center justify-between mt-1">
                        <span className="text-sm font-medium text-gray-700">Password:</span>
                        <span className="text-sm text-gray-900">{"•".repeat(8)}</span>
                      </div>
                    </>
                  )}
                </div>
              </div>

              {proxy.server && (
                <div className="mt-3 pt-3 border-t border-gray-200">
                  <p className="text-sm text-gray-600">
                    Server: <span className="font-medium">{proxy.server.name || 'Unknown'}</span>
                  </p>
                </div>
              )}

              <div className="mt-4 flex justify-between text-xs text-gray-500">
                <span>Created: {new Date(proxy.created_at).toLocaleDateString()}</span>
                <span>Updated: {new Date(proxy.updated_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {(!proxies || proxies.length === 0) && (
        <div className="text-center py-12">
          <p className="text-gray-500">No proxies configured yet</p>
        </div>
      )}
    </div>
  );
};
