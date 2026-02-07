import { type ReactNode } from 'react';
import { Toaster } from 'sonner';
import { Header } from './Header';
import { Footer } from './Footer';

export interface LayoutProps {
  children: ReactNode;
  showHeader?: boolean;
  showAdminLink?: boolean;
  showFooter?: boolean;
}

export function Layout({ children, showHeader = true, showAdminLink = false, showFooter = true }: LayoutProps) {
  return (
    <div className="min-h-screen bg-background flex flex-col">
      {showHeader && <Header showAdminLink={showAdminLink} />}
      <main className="flex-1">{children}</main>
      {showFooter && <Footer />}
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
