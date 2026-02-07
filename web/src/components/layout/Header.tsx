import { Link } from '@tanstack/react-router';
import { Box, Settings } from 'lucide-react';
import { cn } from '@/lib/utils';
import { usePublicStatus } from '@/hooks/useAdmin';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

export interface HeaderProps {
  showAdminLink?: boolean;
  className?: string;
}

export function Header({ showAdminLink = true, className }: HeaderProps) {
  const { data: status } = usePublicStatus();

  return (
    <header className={cn('border-b border-border bg-background/80 backdrop-blur-sm sticky top-0 z-40', className)}>
      <div className="container mx-auto px-4 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <Box className="h-5 w-5 text-background" />
          </div>
          <span className="font-semibold text-lg text-text-primary">WhatsBox</span>
        </Link>

        {/* Right side */}
        <div className="flex items-center gap-4">
          {/* Theme Toggle */}
          <ThemeToggle size="sm" />

          {/* Connection status */}
          <div className="flex items-center gap-2">
            <div
              className={cn(
                'w-2 h-2 rounded-full',
                status?.connected ? 'bg-accent' : 'bg-error'
              )}
            />
            <span className="text-sm text-text-secondary hidden sm:inline">
              {status?.connected ? 'Ready' : 'Offline'}
            </span>
          </div>

          {/* Admin link */}
          {showAdminLink && (
            <Link
              to="/admin"
              className="flex items-center gap-2 text-text-secondary hover:text-text-primary transition-colors"
            >
              <Settings className="h-4 w-4" />
              <span className="hidden sm:inline text-sm">Admin</span>
            </Link>
          )}
        </div>
      </div>
    </header>
  );
}
