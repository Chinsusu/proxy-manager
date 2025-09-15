import React from 'react';
import { useServers } from '../hooks/useServers';

export const Servers: React.FC = () => {
  const { servers, isLoading, error } = useServers();

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
        <p className="text-red-600">Error loading servers</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">Servers</h1>
        <p className="text-gray-600">Manage your proxy servers</p>
      </div>

      <div className="bg-white shadow overflow-hidden sm:rounded-md">
        <ul className="divide-y divide-gray-200">
          {servers?.map((server) => (
            <li key={server.id} className="px-6 py-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center">
                    <h3 className="text-lg font-medium text-gray-900">{server.name}</h3>
                    <span className={`ml-3 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      server.status === 'online' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                    }`}>
                      {server.status}
                    </span>
                  </div>
                  <div className="mt-1">
                    <p className="text-sm text-gray-600">
                      Config Version: v{server.config_version} | 
                      WAN: {server.wan_iface} | 
                      LAN: {server.lan_iface}
                    </p>
                    {server.tags && (
                      <div className="mt-2 flex flex-wrap gap-1">
                        {JSON.parse(server.tags).map((tag: string, index: number) => (
                          <span
                            key={index}
                            className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                <div className="flex items-center space-x-4">
                  <div className="text-sm text-gray-500">
                    {server.proxies && (
                      <span>{server.proxies.length} proxies</span>
                    )}
                  </div>
                  <div className="text-sm text-gray-500">
                    Last seen: {server.last_seen_at ? new Date(server.last_seen_at).toLocaleString() : 'Never'}
                  </div>
                </div>
              </div>
              
              {server.proxies && server.proxies.length > 0 && (
                <div className="mt-4 pl-4 border-l-2 border-gray-200">
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Proxies:</h4>
                  <div className="space-y-2">
                    {server.proxies.map((proxy) => (
                      <div key={proxy.id} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                        <div>
                          <span className="font-medium text-gray-900">{proxy.label}</span>
                          <span className="ml-2 text-sm text-gray-600">
                            {proxy.type.toUpperCase()} - {proxy.host}:{proxy.port}
                          </span>
                        </div>
                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                          proxy.health === 'ok' ? 'bg-green-100 text-green-800' : 
                          proxy.health === 'fail' ? 'bg-red-100 text-red-800' : 'bg-yellow-100 text-yellow-800'
                        }`}>
                          {proxy.health}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};
