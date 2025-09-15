# UI Placeholder

SPA (React/Vue/Angular) build thành static files để Nginx phục vụ từ `/usr/share/nginx/html`.

Yêu cầu chính:
- Login form với email/password → gọi POST `/api/v1/auth/login`
- Lưu JWT token vào localStorage, auto refresh nếu 401
- Pages: Dashboard, Servers, Proxies, Mappings
- CRUD với modal forms (đóng modal sau submit + refresh list)
- Badge trạng thái server (online/offline dựa trên last_seen_at)
- Confirm dangerous actions (delete)
- Toast success/error messages

Tech stack gợi ý:
- React với TanStack Query hoặc SWR
- UI library: Ant Design / Material-UI / Tailwind
- Forms: React Hook Form / Formik
- Toast: react-hot-toast
- Routing: React Router với protected routes
- API calls relative path `/api/v1/...` (không hardcode hostname)

Build process nên output static files và copy vào `ui_build` volume của docker-compose.yml.
