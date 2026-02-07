import { createRouter, createRootRoute, createRoute, Outlet, redirect } from '@tanstack/react-router';
import { Layout, AdminLayout } from '@/components/layout';

// Pages
import Home from '@/pages/Home';
import Download from '@/pages/Download';
import NotFound from '@/pages/NotFound';
import AdminLogin from '@/pages/admin/Login';
import AdminDashboard from '@/pages/admin/Dashboard';
import AdminFiles from '@/pages/admin/Files';
import AdminSettings from '@/pages/admin/Settings';

// Root route
const rootRoute = createRootRoute({
  component: () => <Outlet />,
});

// Public layout route
const publicLayout = createRoute({
  getParentRoute: () => rootRoute,
  id: 'public',
  component: () => (
    <Layout>
      <Outlet />
    </Layout>
  ),
});

// Home route
const homeRoute = createRoute({
  getParentRoute: () => publicLayout,
  path: '/',
  component: Home,
});

// Download route
const downloadRoute = createRoute({
  getParentRoute: () => publicLayout,
  path: '/f/$fileId',
  component: Download,
});

// 404 route
const notFoundRoute = createRoute({
  getParentRoute: () => publicLayout,
  path: '*',
  component: NotFound,
});

// Admin login route (no sidebar)
const adminLoginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin/login',
  component: AdminLogin,
});

// Admin layout route with auth check
const adminLayout = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin',
  component: () => (
    <AdminLayout>
      <Outlet />
    </AdminLayout>
  ),
  pendingComponent: () => (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        <p className="text-text-secondary">Checking authentication...</p>
      </div>
    </div>
  ),
  beforeLoad: async () => {
    // Check auth before loading admin routes
    try {
      const response = await fetch('/api/admin/me', { credentials: 'include' });
      if (!response.ok) {
        throw redirect({ to: '/admin/login' });
      }
    } catch (error) {
      // If it's already a redirect, re-throw it
      if (error && typeof error === 'object' && 'to' in error) {
        throw error;
      }
      throw redirect({ to: '/admin/login' });
    }
  },
});

// Admin dashboard route
const adminDashboardRoute = createRoute({
  getParentRoute: () => adminLayout,
  path: '/',
  component: AdminDashboard,
});

// Admin files route
const adminFilesRoute = createRoute({
  getParentRoute: () => adminLayout,
  path: '/files',
  component: AdminFiles,
});

// Admin settings route
const adminSettingsRoute = createRoute({
  getParentRoute: () => adminLayout,
  path: '/settings',
  component: AdminSettings,
});

// Build route tree
const routeTree = rootRoute.addChildren([
  publicLayout.addChildren([homeRoute, downloadRoute, notFoundRoute]),
  adminLoginRoute,
  adminLayout.addChildren([adminDashboardRoute, adminFilesRoute, adminSettingsRoute]),
]);

// Create router
export const router = createRouter({
  routeTree,
  defaultPreload: 'intent',
});

// Export route for use in components
export { downloadRoute };

// Type declarations
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
