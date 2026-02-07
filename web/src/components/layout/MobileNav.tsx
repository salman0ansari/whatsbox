import { Link, useLocation } from '@tanstack/react-router';
import { LayoutDashboard, Files, Settings, Menu, X, LogOut, Box } from 'lucide-react';
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { useStatus, useLogoutSession } from '@/hooks/useAdmin';
import { Button, ThemeToggle } from '@/components/ui';

const navItems = [
  { to: '/admin', icon: LayoutDashboard, label: 'Dashboard', exact: true },
  { to: '/admin/files', icon: Files, label: 'Files', exact: false },
  { to: '/admin/settings', icon: Settings, label: 'Settings', exact: false },
];

interface MobileNavProps {
  className?: string;
}

export function MobileNav({ className }: MobileNavProps) {
  const [isOpen, setIsOpen] = useState(false);
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
    <>
      {/* Mobile Header */}
      <div className={cn(
        "fixed top-0 left-0 right-0 h-16 border-b border-border bg-surface/95 backdrop-blur-sm z-40 flex items-center justify-between px-4",
        className
      )}>
        <Link to="/" className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <Box className="h-5 w-5 text-background" />
          </div>
          <span className="font-semibold text-lg text-text-primary">WhatsBox</span>
        </Link>
        
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="p-2 rounded-lg hover:bg-surface-hover transition-colors"
        >
          {isOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
        </button>
      </div>

      {/* Mobile Menu Overlay */}
      {isOpen && (
        <div className="fixed inset-0 z-30 lg:hidden">
          {/* Backdrop */}
          <div 
            className="absolute inset-0 bg-black/50 backdrop-blur-sm"
            onClick={() => setIsOpen(false)}
          />
          
          {/* Menu Panel */}
          <div className="absolute top-16 left-0 right-0 bg-surface border-b border-border shadow-xl">
            <nav className="p-4 space-y-1">
              {navItems.map((item) => (
                <Link
                  key={item.to}
                  to={item.to}
                  onClick={() => setIsOpen(false)}
                  className={cn(
                    'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
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

            {/* Mobile Menu Footer */}
            <div className="p-4 border-t border-border space-y-3">
              {/* Theme Toggle */}
              <div className="flex items-center justify-between px-4 py-2">
                <span className="text-sm text-text-secondary">Theme</span>
                <ThemeToggle size="sm" />
              </div>

              {/* Connection Status */}
              <div className="flex items-center gap-2 px-4 py-2">
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
                onClick={() => {
                  logoutMutation.mutate();
                  setIsOpen(false);
                }}
                loading={logoutMutation.isPending}
              >
                <LogOut className="h-4 w-4 mr-2" />
                Logout
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Bottom Navigation Bar */}
      <div className={cn(
        "fixed bottom-0 left-0 right-0 h-16 border-t border-border bg-surface z-40 lg:hidden",
        className
      )}>
        <nav className="flex items-center justify-around h-full">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={cn(
                'flex flex-col items-center justify-center gap-1 px-4 py-2 rounded-lg transition-colors',
                isActive(item.to, item.exact)
                  ? 'text-accent'
                  : 'text-text-secondary'
              )}
            >
              <item.icon className="h-5 w-5" />
              <span className="text-xs">{item.label}</span>
            </Link>
          ))}
        </nav>
      </div>
    </>
  );
}
