import { Link } from '@tanstack/react-router';
import { FileQuestion } from 'lucide-react';
import { Button } from '@/components/ui';

export default function NotFound() {
  return (
    <div className="container mx-auto px-4 py-20 max-w-md text-center">
      <div className="w-16 h-16 rounded-full bg-surface-hover flex items-center justify-center mx-auto mb-6">
        <FileQuestion className="h-8 w-8 text-text-secondary" />
      </div>
      <h1 className="text-2xl font-bold text-text-primary mb-2">
        Page Not Found
      </h1>
      <p className="text-text-secondary mb-6">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link to="/">
        <Button>Go Home</Button>
      </Link>
    </div>
  );
}
