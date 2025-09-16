# Proxy Group Management

## Overview
The Proxy Group Management feature allows you to organize proxies into logical groups for better management and organization. This feature includes full CRUD operations for groups and the ability to move proxies between groups.

## Features

### Group Management
- **Create Groups**: Add new proxy groups with name and description
- **Edit Groups**: Update existing group information
- **Delete Groups**: Remove empty groups (groups with proxies cannot be deleted)
- **View Groups**: See all groups with proxy counts

### Proxy Organization
- **Move Single Proxy**: Move individual proxies to different groups
- **Bulk Move Proxies**: Move multiple selected proxies to a group at once
- **Group View**: Toggle between viewing proxies by groups or by servers
- **No Group**: Proxies can exist without a group assignment

## API Endpoints

### Group Operations
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/groups` | List all groups | Yes |
| POST | `/api/v1/groups` | Create new group | Yes |
| PUT | `/api/v1/groups/:id` | Update group | Yes |
| DELETE | `/api/v1/groups/:id` | Delete group | Yes |

### Proxy Move Operations
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| PUT | `/api/v1/proxies/:id/group` | Move single proxy to group | Yes |
| PUT | `/api/v1/proxies/bulk-move` | Move multiple proxies to group | Yes |

## API Request/Response Examples

### Create Group
**Request:**
```json
POST /api/v1/groups
{
  "name": "Production Proxies",
  "description": "Proxies for production environment"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "Production Proxies",
  "description": "Proxies for production environment",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Move Single Proxy
**Request:**
```json
PUT /api/v1/proxies/123/group
{
  "group_id": 5
}
```

**Response:**
```json
{
  "id": 123,
  "label": "Proxy-1",
  "group_id": 5,
  ...
}
```

### Bulk Move Proxies
**Request:**
```json
PUT /api/v1/proxies/bulk-move
{
  "proxy_ids": [1, 2, 3, 4],
  "group_id": 5
}
```

**Response:**
```json
{
  "message": "Proxies moved successfully",
  "moved_count": 4
}
```

## Database Schema

### proxy_groups Table
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| name | VARCHAR | Unique group name |
| description | TEXT | Optional description |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

### proxies Table (Modified)
| Column | Type | Description |
|--------|------|-------------|
| ... | ... | (existing columns) |
| group_id | INTEGER | Foreign key to proxy_groups.id (nullable) |

## UI Components

### GroupManagement Component
- Located in `ui/src/components/GroupManagement.tsx`
- Provides full CRUD interface for managing groups
- Uses React Query for state management
- Real-time updates with toast notifications

### MoveProxyButton Component
- Located in `ui/src/components/MoveProxyButton.tsx`
- Dropdown selector for moving single proxy
- Shows all available groups including "No Group" option

### BulkMoveProxyButton Component
- Located in `ui/src/components/BulkMoveProxyButton.tsx`
- Modal interface for bulk moving selected proxies
- Only visible when proxies are selected

## Usage Guide

### Creating a Group
1. Navigate to the Proxies page
2. Toggle to "Group View" using the switch button
3. Click "New Group" button
4. Enter group name and optional description
5. Click "Create"

### Moving Proxies to Groups

#### Single Proxy
1. Find the proxy you want to move
2. Click the dropdown next to the proxy
3. Select the target group
4. Click "Move" button

#### Multiple Proxies
1. Select checkboxes for proxies to move
2. Click "Bulk Move" button that appears
3. Select target group in the modal
4. Click "Move X Proxies"

### Deleting a Group
1. Groups can only be deleted when empty
2. Move all proxies out of the group first
3. Click the trash icon next to the group
4. Confirm deletion

## Technical Implementation

### Backend (Go)
- **Models**: Added `ProxyGroup` struct in `models.go`
- **Handlers**: Created `groups.go` handler with CRUD operations
- **Routes**: Registered group endpoints in `main.go`
- **Database**: Auto-migration creates tables and relationships

### Frontend (React/TypeScript)
- **State Management**: React Query for caching and mutations
- **API Integration**: Axios with JWT authentication
- **UI Updates**: Optimistic updates with cache invalidation
- **Notifications**: Toast messages for user feedback

## Environment Variables
No additional environment variables required. The feature uses existing database and API configurations.

## Migration Notes
- Database migration runs automatically on API startup
- Existing proxies will have `group_id` set to NULL
- No data loss during migration

## Security Considerations
- All group operations require JWT authentication
- Groups are tenant-isolated (if multi-tenancy is implemented)
- Input validation on both frontend and backend
- SQL injection prevention through ORM

## Performance Considerations
- Groups are eagerly loaded with proxy counts
- Bulk operations use single database transaction
- Efficient queries with proper indexes
- React Query caching reduces API calls

## Future Enhancements
- Group-based access control
- Nested groups/sub-groups
- Group templates
- Bulk operations at group level
- Group-specific configurations
- Export/Import groups
