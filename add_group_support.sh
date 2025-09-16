#!/bin/bash

echo "Adding group support to Proxies.tsx..."

# 1. Add ProxyGroup interface
sed -i '/import { BulkDeleteProxyButton } from.*$/a\\n\
interface ProxyGroup {\
  id: string;\
  name: string;\
  description?: string;\
  color?: string;\
  created_at: string;\
  updated_at: string;\
}' ui/src/pages/Proxies.tsx

# 2. Add group_id to Proxy interface
sed -i '/server_id: string;/a\
  group_id?: string;' ui/src/pages/Proxies.tsx

# 3. Import new components  
sed -i '/import { BulkDeleteProxyButton } from.*$/a\
import { GroupManagement } from '\''../components/GroupManagement'\'';\
import { MoveProxyButton } from '\''../components/MoveProxyButton'\'';' ui/src/pages/Proxies.tsx

# 4. Add showGroupView state
sed -i '/const \[selectedProxies, setSelectedProxies\] = useState/a\
  const [showGroupView, setShowGroupView] = useState<boolean>(false);' ui/src/pages/Proxies.tsx

echo "Group support added successfully!"
