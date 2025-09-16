#!/bin/bash

# Fix 1: Update ExistingProxy interface in ProxyBulkImportModal.tsx
sed -i '/interface ExistingProxy {/,/}/ {
  /}/i\
  username?: string;\
  password?: string;
}' ui/src/components/ProxyBulkImportModal.tsx

# Fix 2: Update detectDuplicates function in ProxyBulkImportModal.tsx  
sed -i '/existing => existing.host === proxy.host && existing.port === proxy.port/c\
        existing => existing.host === proxy.host && \
                   existing.port === proxy.port &&\
                   (existing.username || "") === proxy.username &&\
                   (existing.password || "") === proxy.password' ui/src/components/ProxyBulkImportModal.tsx

# Fix 3: Update getExistingProxies function in Proxies.tsx
sed -i '/label: proxy.label$/c\
        label: proxy.label,\
        username: proxy.username,\
        password: proxy.password' ui/src/pages/Proxies.tsx

echo "Fixed duplicate detection logic!"
