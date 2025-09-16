# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2024-09-17

### Added
- **Proxy Group Management System**
  - Create, read, update, delete proxy groups
  - Move single proxy between groups
  - Bulk move multiple proxies to groups
  - Group view toggle in UI
  - Group selector in proxy import modal
  
- **Backend Enhancements**
  - `ProxyGroup` model with CRUD operations
  - `group_id` field in Proxy model
  - Group management handler (`groups.go`)
  - Move proxy API endpoints
  - Database auto-migration for groups
  
- **Frontend Components**
  - `GroupManagement` component for group CRUD
  - `MoveProxyButton` for single proxy moves
  - `BulkMoveProxyButton` for bulk operations
  - Group view/Server view toggle
  
- **Documentation**
  - Complete proxy groups guide (`docs/PROXY_GROUPS.md`)
  - Updated API documentation with group endpoints
  - Enhanced README with feature descriptions

### Changed
- Proxies page now supports dual view modes (groups/servers)
- Database schema updated with proxy_groups table
- API routes updated to include group endpoints
- Frontend uses real API calls instead of mock data

### Fixed
- Persistent data storage - no more data loss on refresh
- React Query cache management for instant UI updates
- TypeScript interface compatibility issues

## [1.1.0] - 2024-09-16

### Fixed
- Proxy duplicate detection logic
- Bulk import validation improvements
- Import modal duplicate handling

### Changed
- Duplicate detection now matches backend logic
- Auto-skip duplicates during import

## [1.0.0] - 2024-09-15

### Added
- Initial release
- Server management (CRUD operations)
- Proxy management with health monitoring
- Mapping rules configuration
- Agent pull configuration API
- JWT authentication
- Admin user management
- Docker compose setup
- Nginx reverse proxy
- PostgreSQL database
- React UI with TypeScript
- Cloudflare Tunnel support

### Features
- Real-time health monitoring
- Versioned configuration for agents
- Bulk proxy import
- Password visibility toggle
- Responsive UI design
- Toast notifications
- Loading states
- Error handling

## [0.1.0] - 2024-09-14

### Added
- Project initialization
- Basic infrastructure setup
- Documentation structure
