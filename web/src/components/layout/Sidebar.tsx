import { Link, useLocation } from '@tanstack/react-router';
import { LayoutDashboard, Files, Settings, LogOut, Box } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useStatus, useLogoutSession } from '@/hooks/useAdmin';
import { Button } from '@/components/ui';

const navItems = [
  { to: '/admin', icon: LayoutDashboard, label: 'Dashboard', exact: true },
  { to: '/admin/files', icon: Files, label: 'Files', exact: false },
  { to: '/admin/settings', icon: Settings, label: 'Settings', exact: false },
];

export function Sidebar() {
  const location = useLocation();
  const { data: status } = useStatus();
  const logoutMutation = useLogoutSession();

  const isActive = (to: string, exact: boolean) => {
    if (exact) {
      return location.pathname === to;
    }
    return location.pathname.startsWith(to);
  };

  return (
    <aside className="w-64 border-r border-border bg-surface min-h-screen flex flex-col">
      {/* Logo */}
      <div className="h-16 border-b border-border flex items-center px-4">
        <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <Box className="h-5 w-5 text-background" />
          </div>
          <span className="font-semibold text-lg text-text-primary">WhatsBox</span>
        </Link>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => (
          <Link
            key={item.to}
            to={item.to}
            className={cn(
              'flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors',
              isActive(item.to, item.exact)
                ? 'bg-accent/10 text-accent'
                : 'text-text-secondary hover:text-text-primary hover:bg-surface-hover'
            )}
          >
            <item.icon className="h-5 w-5" />
            <span className="font-medium">{item.label}</span>
          </Link>
        ))}
      </nav>

      {/* Status & Logout */}
      <div className="p-4 border-t border-border space-y-3">
        {/* Connection Status */}
        <div className="flex items-center gap-2 px-3 py-2">
          <div
            className={cn(
              'w-2 h-2 rounded-full',
              status?.connected ? 'bg-accent' : 'bg-error'
            )}
          />
          <span className="text-sm text-text-secondary">
            {status?.connected ? 'WhatsApp Connected' : 'Disconnected'}
          </span>
        </div>

        {/* Logout Button */}
        <Button
          variant="ghost"
          className="w-full justify-start text-text-secondary hover:text-error"
          onClick={() => logoutMutation.mutate()}
          loading={logoutMutation.isPending}
        >
          <LogOut className="h-4 w-4 mr-2" />
          Logout
        </Button>
      </div>
    </aside>
  );
}
