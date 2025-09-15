import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { useAuth } from './hooks/useAuth';
import { Layout } from './components/Layout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
    },
  },
});

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuth();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

const AppContent: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/*"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route 
            path="servers" 
            element={
              <div className="text-center py-12">
                <h2 className="text-2xl font-bold text-gray-900">Servers</h2>
                <p className="text-gray-600">Server management coming soon...</p>
              </div>
            } 
          />
          <Route 
            path="proxies" 
            element={
              <div className="text-center py-12">
                <h2 className="text-2xl font-bold text-gray-900">Proxies</h2>
                <p className="text-gray-600">Proxy management coming soon...</p>
              </div>
            } 
          />
          <Route 
            path="mappings" 
            element={
              <div className="text-center py-12">
                <h2 className="text-2xl font-bold text-gray-900">Mappings</h2>
                <p className="text-gray-600">Mapping management coming soon...</p>
              </div>
            } 
          />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppContent />
      <Toaster position="top-right" />
    </QueryClientProvider>
  );
}

export default App;
