import React, { useState } from 'react';
import { X, Upload, FileText, CheckCircle, XCircle, Loader, AlertTriangle } from 'lucide-react';

interface ProxyData {
  host: string;
  port: number;
  username: string;
  password: string;
  type?: string;
  health?: string;
  label?: string;
}

interface ProxyBulkImportModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImport: (proxies: ProxyData[]) => void;
  isLoading?: boolean;
  availableServers: Array<{ id: string; name: string }>;
}

export const ProxyBulkImportModal: React.FC<ProxyBulkImportModalProps> = ({
  isOpen,
  onClose,
  onImport,
  isLoading = false,
  availableServers
}) => {
  const [proxyText, setProxyText] = useState('');
  const [selectedServerId, setSelectedServerId] = useState<string>('');
  const [parsedProxies, setParsedProxies] = useState<ProxyData[]>([]);
  const [isChecking, setIsChecking] = useState(false);
  const [step, setStep] = useState<'input' | 'checking' | 'results'>('input');

  // Sample proxies for demonstration
  const sampleProxies = `176.46.132.43:45941:cCP3nka9oeo5jKz:4XmT66DrXM6bPuR
46.203.160.111:45911:6duP8kjjcf9dbyu:w5EBeSV81S9QBuO
46.203.160.93:47101:0u4M88sgefKLAWr:oZnoaOOotxSafhV`;

  const parseProxyList = (text: string): ProxyData[] => {
    const lines = text.trim().split('\n');
    const proxies: ProxyData[] = [];

    lines.forEach((line, index) => {
      const trimmed = line.trim();
      if (!trimmed) return;

      const parts = trimmed.split(':');
      if (parts.length >= 4) {
        const host = parts[0];
        const port = parseInt(parts[1]);
        const username = parts[2];
        const password = parts.slice(3).join(':'); // Handle passwords with colons

        if (host && !isNaN(port) && username && password) {
          proxies.push({
            host,
            port,
            username,
            password,
            label: `Proxy-${host.split('.').pop()}-${port}`,
            type: 'unknown',
            health: 'unknown'
          });
        }
      }
    });

    return proxies;
  };

  const checkProxyType = async (proxy: ProxyData): Promise<{ type: string; health: string }> => {
    // Simulate proxy checking - in real implementation, this would test connectivity
    return new Promise((resolve) => {
      setTimeout(() => {
        // Simulate different proxy types based on port patterns
        let type = 'socks5';
        let health = Math.random() > 0.2 ? 'ok' : 'fail'; // 80% success rate

        // Common HTTP proxy ports
        if ([8080, 3128, 8888, 80, 443].includes(proxy.port)) {
          type = 'http';
        }
        // Common SOCKS proxy ports  
        else if ([1080, 9050, 9051].includes(proxy.port)) {
          type = 'socks5';
        }
        // For other ports, randomly assign but prefer socks5
        else {
          type = Math.random() > 0.3 ? 'socks5' : 'http';
        }

        resolve({ type, health });
      }, Math.random() * 2000 + 500); // Random delay 500-2500ms
    });
  };

  const handleParseAndCheck = async () => {
    const proxies = parseProxyList(proxyText);
    if (proxies.length === 0) {
      alert('No valid proxies found. Please check the format: ip:port:user:pass');
      return;
    }

    setParsedProxies(proxies);
    setStep('checking');
    setIsChecking(true);

    // Check each proxy type and health
    const checkedProxies: ProxyData[] = [];
    for (const proxy of proxies) {
      const { type, health } = await checkProxyType(proxy);
      checkedProxies.push({
        ...proxy,
        type,
        health
      });
      setParsedProxies([...checkedProxies]); // Update UI progressively
    }

    setIsChecking(false);
    setStep('results');
  };

  const handleImport = () => {
    const validProxies = parsedProxies.filter(p => p.health === 'ok');
    if (validProxies.length === 0) {
      alert('No working proxies found to import');
      return;
    }
    onImport(validProxies);
  };

  const handleReset = () => {
    setProxyText('');
    setParsedProxies([]);
    setStep('input');
    setIsChecking(false);
  };

  const loadSample = () => {
    setProxyText(sampleProxies);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-10 mx-auto p-5 border w-[700px] shadow-lg rounded-md bg-white max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <Upload className="h-5 w-5 text-blue-500" />
            <h3 className="text-lg font-medium text-gray-900">Bulk Import Proxies</h3>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 transition-colors">
            <X className="h-5 w-5" />
          </button>
        </div>

        {step === 'input' && (
          <div className="space-y-4">
            <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
              <p className="text-sm text-blue-800">
                <strong>Format:</strong> Each proxy on a new line in format: <code>ip:port:username:password</code>
              </p>
            </div>

            <div>
              <div className="flex justify-between items-center mb-2">
                <label className="block text-sm font-medium text-gray-700">
                  Proxy List
                </label>
                <button
                  onClick={loadSample}
                  className="text-sm text-blue-600 hover:text-blue-800 transition-colors"
                >
                  Load Sample
                </button>
              </div>
              <textarea
                value={proxyText}
                onChange={(e) => setProxyText(e.target.value)}
                className="w-full h-64 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
                placeholder="176.46.132.43:45941:cCP3nka9oeo5jKz:4XmT66DrXM6bPuR&#10;46.203.160.111:45911:6duP8kjjcf9dbyu:w5EBeSV81S9QBuO&#10;46.203.160.93:47101:0u4M88sgefKLAWr:oZnoaOOotxSafhV"
                disabled={isLoading}
              />
            </div>

            <div className="flex justify-end space-x-3">
              <button
                onClick={onClose}
                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
                disabled={isLoading}
              >
                Cancel
              </button>
              <button
                onClick={handleParseAndCheck}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 transition-colors"
                disabled={!proxyText.trim() || isLoading}
              >
                Parse & Check Proxies
              </button>
            </div>
          </div>
        )}

        {step === 'checking' && (
          <div className="space-y-4">
            <div className="text-center py-8">
              <Loader className="h-8 w-8 animate-spin mx-auto mb-4 text-blue-500" />
              <h4 className="text-lg font-medium text-gray-900 mb-2">Checking Proxies</h4>
              <p className="text-gray-600">Testing proxy connectivity and determining types...</p>
            </div>

            {parsedProxies.length > 0 && (
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {parsedProxies.map((proxy, index) => (
                  <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                    <div className="flex items-center space-x-3">
                      <div>
                        {proxy.health === 'unknown' ? (
                          <Loader className="h-4 w-4 animate-spin text-blue-500" />
                        ) : proxy.health === 'ok' ? (
                          <CheckCircle className="h-4 w-4 text-green-500" />
                        ) : (
                          <XCircle className="h-4 w-4 text-red-500" />
                        )}
                      </div>
                      <div>
                        <div className="text-sm font-medium">{proxy.host}:{proxy.port}</div>
                        <div className="text-xs text-gray-500">{proxy.username}</div>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-medium">{proxy.type}</div>
                      <div className="text-xs text-gray-500">{proxy.health}</div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {step === 'results' && (
          <div className="space-y-4">
            <div className="bg-green-50 border border-green-200 rounded-md p-3">
              <div className="flex items-center space-x-2">
                <CheckCircle className="h-5 w-5 text-green-500" />
                <p className="text-sm text-green-800">
                  <strong>Check Complete:</strong> {parsedProxies.filter(p => p.health === 'ok').length} working, {parsedProxies.filter(p => p.health === 'fail').length} failed out of {parsedProxies.length} total proxies
                </p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4 text-sm">
              <div className="text-center p-3 bg-blue-50 rounded">
                <div className="font-medium text-blue-900">HTTP Proxies</div>
                <div className="text-2xl font-bold text-blue-600">
                  {parsedProxies.filter(p => p.type === 'http' && p.health === 'ok').length}
                </div>
              </div>
              <div className="text-center p-3 bg-purple-50 rounded">
                <div className="font-medium text-purple-900">SOCKS5 Proxies</div>
                <div className="text-2xl font-bold text-purple-600">
                  {parsedProxies.filter(p => p.type === 'socks5' && p.health === 'ok').length}
                </div>
              </div>
            </div>

            <div className="space-y-2 max-h-64 overflow-y-auto">
              {parsedProxies.map((proxy, index) => (
                <div 
                  key={index} 
                  className={`flex items-center justify-between p-3 rounded ${
                    proxy.health === 'ok' ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'
                  }`}
                >
                  <div className="flex items-center space-x-3">
                    <div>
                      {proxy.health === 'ok' ? (
                        <CheckCircle className="h-4 w-4 text-green-500" />
                      ) : (
                        <XCircle className="h-4 w-4 text-red-500" />
                      )}
                    </div>
                    <div>
                      <div className="text-sm font-medium">{proxy.host}:{proxy.port}</div>
                      <div className="text-xs text-gray-500">{proxy.username} • {proxy.label}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className={`text-sm font-medium px-2 py-1 rounded text-white ${
                      proxy.type === 'http' ? 'bg-blue-500' : 'bg-purple-500'
                    }`}>
                      {(proxy.type || "UNKNOWN").toUpperCase()}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">{proxy.health}</div>
                  </div>
                </div>
              ))}
            </div>

            <div className="flex justify-between space-x-3">
              <button
                onClick={handleReset}
                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
              >
                Start Over
              </button>
              <div className="space-x-3">
                <button
                  onClick={onClose}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleImport}
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 transition-colors disabled:opacity-50"
                  disabled={parsedProxies.filter(p => p.health === 'ok').length === 0 || isLoading}
                >
                  {isLoading ? 'Importing...' : `Import ${parsedProxies.filter(p => p.health === 'ok').length} Working Proxies`}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
