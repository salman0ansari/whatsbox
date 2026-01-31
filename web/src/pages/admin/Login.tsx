import { useState } from 'react';
import { Box, Lock } from 'lucide-react';
import { Card, Button, Input } from '@/components/ui';
import { useLogin } from '@/hooks';

export default function AdminLogin() {
  const [password, setPassword] = useState('');
  const loginMutation = useLogin();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (password) {
      loginMutation.mutate(password);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <Card className="w-full max-w-sm" padding="lg">
        <div className="text-center mb-8">
          <div className="w-12 h-12 rounded-xl bg-accent flex items-center justify-center mx-auto mb-4">
            <Box className="h-7 w-7 text-background" />
          </div>
          <h1 className="text-xl font-bold text-text-primary">
            WhatsBox Admin
          </h1>
          <p className="text-sm text-text-secondary mt-1">
            Enter your password to continue
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            type="password"
            placeholder="Enter admin password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoFocus
          />
          
          <Button
            type="submit"
            className="w-full"
            loading={loginMutation.isPending}
            disabled={!password}
          >
            <Lock className="h-4 w-4" />
            Sign In
          </Button>
        </form>
      </Card>
    </div>
  );
}
