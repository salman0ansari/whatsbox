import { type ReactNode } from 'react';
import { Toaster } from 'sonner';
import { Sidebar } from './Sidebar';

export interface AdminLayoutProps {
  children: ReactNode;
}

export function AdminLayout({ children }: AdminLayoutProps) {
  return (
    <div className="min-h-screen bg-background flex">
      <Sidebar />
      <main className="flex-1 overflow-auto">
        {children}
      </main>
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
