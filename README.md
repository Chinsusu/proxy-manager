# Proxy Manager Control-Plane

A comprehensive proxy management system with web UI and REST API for managing multiple proxy servers. Access via Cloudflare Tunnel at `https://pmu.xelu.top`.

## 🚀 Features

### Core Features
- **Server Management**: Add, edit, delete proxy server nodes/agents
- **Proxy Management**: Manage upstream proxies with health monitoring
- **Mapping Rules**: Configure client-to-proxy routing rules
- **Health Monitoring**: Real-time health checks for proxies and servers
- **Versioned Configuration**: Agent pull configuration with version tracking

### New: Proxy Group Management 🎯
- **Group Organization**: Organize proxies into logical groups
- **CRUD Operations**: Create, read, update, delete proxy groups
- **Move Operations**: Move single or multiple proxies between groups
- **Flexible Views**: Toggle between group view and server view
- **Bulk Actions**: Select and move multiple proxies at once

## 📋 System Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────┐
│   Browser   │────▶│  Nginx:8080  │────▶│ API:8082 │
└─────────────┘     └──────────────┘     └──────────┘
                           │                    │
                    ┌──────▼──────┐      ┌─────▼────┐
                    │  UI (React)  │      │ PostgreSQL│
                    └──────────────┘      └──────────┘
```

## 🛠️ Tech Stack

- **Frontend**: React, TypeScript, Tailwind CSS, React Query
- **Backend**: Go, Gin, GORM, PostgreSQL
- **Infrastructure**: Docker, Nginx, Cloudflare Tunnel
- **Authentication**: JWT

## 📦 Quick Start

### Prerequisites
- Docker & Docker Compose
- Node.js 18+ (for development)
- Go 1.21+ (for development)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Chinsusu/proxy-manager.git
cd proxy-manager-control-plane
```

2. Configure environment:
```bash
cp .env.example .env
# Edit .env file with your settings
```

3. Start services:
```bash
docker-compose up -d
```

4. Access the application:
- Local: http://localhost:8080
- Production: https://pmu.xelu.top

### Default Credentials
- Email: `admin@example.com`
- Password: `admin123`

## 📚 Documentation

- [API Documentation](docs/API_CONTRACT.md) - REST API endpoints and contracts
- [Proxy Groups Guide](docs/PROXY_GROUPS.md) - Complete guide for proxy group management
- [Deployment Guide](docs/DEPLOY.md) - Production deployment instructions
- [Cloudflare Tunnel Setup](docs/CLOUDFLARE_TUNNEL.md) - Tunnel configuration guide

## 🔧 Development

### Backend Development
```bash
cd api
go mod download
go run cmd/server/main.go
```

### Frontend Development
```bash
cd ui
npm install
npm start
```

### Build for Production

#### Backend
```bash
cd api
go build -o api-server cmd/server/main.go
```

#### Frontend
```bash
cd ui
npm run build
```

## 🗂️ Project Structure

```
proxy-manager-control-plane/
├── api/                    # Go backend API
│   ├── cmd/               # Application entrypoints
│   ├── internal/          # Internal packages
│   │   ├── config/       # Configuration
│   │   ├── database/     # Database connection
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── models/       # Database models
│   └── migrations/        # Database migrations
├── ui/                     # React frontend
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── pages/        # Page components
│   │   ├── lib/          # Utilities and API client
│   │   └── types/        # TypeScript types
│   └── build/            # Production build output
├── nginx/                  # Nginx configuration
├── docs/                   # Documentation
└── docker-compose.yml      # Docker composition
```

## 🔐 API Endpoints

### Authentication
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/me` - Get current user

### Servers
- `GET /api/v1/servers` - List servers
- `POST /api/v1/servers` - Create server
- `GET /api/v1/servers/:id` - Get server
- `PATCH /api/v1/servers/:id` - Update server
- `DELETE /api/v1/servers/:id` - Delete server

### Proxies
- `GET /api/v1/proxies` - List proxies
- `POST /api/v1/proxies` - Create proxy
- `GET /api/v1/proxies/:id` - Get proxy
- `PATCH /api/v1/proxies/:id` - Update proxy
- `DELETE /api/v1/proxies/:id` - Delete proxy
- `PUT /api/v1/proxies/:id/group` - Move proxy to group
- `PUT /api/v1/proxies/bulk-move` - Bulk move proxies

### Groups (New)
- `GET /api/v1/groups` - List groups
- `POST /api/v1/groups` - Create group
- `PUT /api/v1/groups/:id` - Update group
- `DELETE /api/v1/groups/:id` - Delete group

### Mappings
- `GET /api/v1/mappings` - List mappings
- `POST /api/v1/mappings` - Create mapping
- `GET /api/v1/mappings/:id` - Get mapping
- `PATCH /api/v1/mappings/:id` - Update mapping
- `DELETE /api/v1/mappings/:id` - Delete mapping

### Agent
- `GET /api/v1/agents/:id/pull` - Pull configuration
- `POST /api/v1/agents/:id/ack` - Acknowledge configuration

## 🚢 Deployment

### Using Docker Compose

```bash
# Build and start all services
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Environment Variables

Key environment variables in `.env`:

```bash
# Database
POSTGRES_USER=pgm
POSTGRES_PASSWORD=your_password
POSTGRES_DB=pgm_db

# API
API_BIND=:8082
API_JWT_SECRET=your_jwt_secret
API_ADMIN_EMAIL=admin@example.com
API_ADMIN_PASSWORD=admin123

# UI
UI_PUBLIC_URL=https://pmu.xelu.top

# Nginx
NGINX_LISTEN=8080
```

## 🧪 Testing

### Test Group Management
1. Create a new group via UI
2. Move proxies to the group
3. Bulk select and move multiple proxies
4. Edit group information
5. Delete empty groups

### API Testing with cURL

```bash
# Login
TOKEN=$(curl -X POST http://localhost:8082/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}' \
  | jq -r '.access_token')

# Create group
curl -X POST http://localhost:8082/api/v1/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Group","description":"Test"}'

# List groups
curl http://localhost:8082/api/v1/groups \
  -H "Authorization: Bearer $TOKEN"
```

## 🐛 Known Issues

- Health check shows "unhealthy" but API works fine (HEAD request issue)
- Initial page load may be slow due to cold start

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- Built with Go and React
- Uses PostgreSQL for data persistence
- Deployed with Docker and Cloudflare Tunnel
- UI components from Tailwind CSS

## 📞 Support

For issues and questions:
- Open an issue on GitHub
- Contact: admin@example.com

## 🔄 Recent Updates

### v1.2.0 - Proxy Group Management
- Added full proxy group CRUD operations
- Implemented single and bulk proxy move features
- Added group view toggle in UI
- Database migration for group support
- Complete API integration with persistent storage

### v1.1.0 - Duplicate Detection Fix
- Fixed proxy duplicate detection logic
- Improved bulk import validation

### v1.0.0 - Initial Release
- Core proxy management features
- Server and mapping management
- Agent pull configuration
