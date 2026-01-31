import { type ReactNode } from 'react';
import { Toaster } from 'sonner';
import { Header } from './Header';

export interface LayoutProps {
  children: ReactNode;
  showHeader?: boolean;
  showAdminLink?: boolean;
}

export function Layout({ children, showHeader = true, showAdminLink = true }: LayoutProps) {
  return (
    <div className="min-h-screen bg-background">
      {showHeader && <Header showAdminLink={showAdminLink} />}
      <main>{children}</main>
      <Toaster
        position="bottom-right"
        toastOptions={{
          style: {
            background: '#141414',
            border: '1px solid #262626',
            color: '#fafafa',
          },
        }}
      />
    </div>
  );
}
