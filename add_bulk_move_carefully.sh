#!/bin/bash

echo "Adding bulk move functionality carefully..."

# 1. Add import for BulkMoveProxyButton
sed -i '/import { MoveProxyButton } from.*$/a\
import { BulkMoveProxyButton } from '\''../components/BulkMoveProxyButton'\'';' ui/src/pages/Proxies.tsx

# 2. Add helper function for getting groups
sed -i '/const clearSelection = () => {/i\
  // Get available groups from cache for bulk operations\
  const getAvailableGroups = () => {\
    const groups = queryClient.getQueryData<any[]>(['\''groups'\'']) || [];\
    return groups.length > 0 ? groups : [{id: '\''1'\'', name: '\''Default'\'', color: '\''#3B82F6'\''}];\
  };\
' ui/src/pages/Proxies.tsx

echo "Added bulk move support!"
