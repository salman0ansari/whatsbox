import { type ReactNode } from 'react';
import { Toaster } from 'sonner';
import { Sidebar } from './Sidebar';
import { MobileNav } from './MobileNav';

export interface AdminLayoutProps {
  children: ReactNode;
}

export function AdminLayout({ children }: AdminLayoutProps) {
  return (
    <div className="min-h-screen bg-background flex">
      <Sidebar className="hidden lg:flex" />
      <main className="flex-1 overflow-auto pb-20 lg:pb-0">
        {children}
      </main>
      <MobileNav className="lg:hidden" />
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
