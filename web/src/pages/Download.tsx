import { useState, useEffect, useCallback } from 'react';
import { useParams } from '@tanstack/react-router';
import { Download as DownloadIcon, Lock, AlertCircle, CheckCircle } from 'lucide-react';
import { Card, Button, Input, Spinner } from '@/components/ui';
import { useFile } from '@/hooks';
import { downloadFile, getDownloadUrl } from '@/api/files';
import { getErrorMessage } from '@/api/client';
import { formatBytes } from '@/lib/utils';
import { downloadRoute } from '@/routes';

export default function Download() {
  const { fileId } = useParams({ from: downloadRoute.id });
  const { data: file, isLoading, error } = useFile(fileId);
  
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [downloading, setDownloading] = useState(false);
  const [downloadStarted, setDownloadStarted] = useState(false);

  const triggerDownload = useCallback(async (pwd?: string) => {
    if (!file) return;
    
    setDownloading(true);
    setPasswordError('');
    
    try {
      if (file.password_protected && pwd) {
        // Download with password via API to validate
        const blob = await downloadFile(file.id, pwd);
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = file.filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      } else {
        // Direct download via browser
        const a = document.createElement('a');
        a.href = getDownloadUrl(file.id);
        a.download = file.filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
      }
      setDownloadStarted(true);
    } catch (error) {
      const message = getErrorMessage(error);
      if (message.toLowerCase().includes('password')) {
        setPasswordError('Incorrect password');
      } else {
        setPasswordError(message);
      }
    } finally {
      setDownloading(false);
    }
  }, [file]);

  // Auto-download for non-password-protected files
  useEffect(() => {
    if (file && !file.password_protected && file.status === 'active' && !downloadStarted) {
      triggerDownload();
    }
  }, [file, downloadStarted, triggerDownload]);

  const handlePasswordSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (password) {
      triggerDownload(password);
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-20 max-w-md">
        <Card className="text-center py-12">
          <Spinner size="lg" className="mx-auto" />
          <p className="mt-4 text-text-secondary">Loading file info...</p>
        </Card>
      </div>
    );
  }

  // Error state
  if (error || !file) {
    return (
      <div className="container mx-auto px-4 py-20 max-w-md">
        <Card className="text-center py-12">
          <div className="w-12 h-12 rounded-full bg-error/20 flex items-center justify-center mx-auto mb-4">
            <AlertCircle className="h-6 w-6 text-error" />
          </div>
          <h2 className="text-lg font-semibold text-text-primary mb-2">
            File Not Found
          </h2>
          <p className="text-text-secondary">
            This file may have been deleted or the link is invalid.
          </p>
        </Card>
      </div>
    );
  }

  // Expired file
  if (file.status === 'expired') {
    return (
      <div className="container mx-auto px-4 py-20 max-w-md">
        <Card className="text-center py-12">
          <div className="w-12 h-12 rounded-full bg-warning/20 flex items-center justify-center mx-auto mb-4">
            <AlertCircle className="h-6 w-6 text-warning" />
          </div>
          <h2 className="text-lg font-semibold text-text-primary mb-2">
            File Expired
          </h2>
          <p className="text-text-secondary">
            This file has expired and is no longer available for download.
          </p>
        </Card>
      </div>
    );
  }

  // Max downloads reached
  if (file.max_downloads && file.download_count >= file.max_downloads) {
    return (
      <div className="container mx-auto px-4 py-20 max-w-md">
        <Card className="text-center py-12">
          <div className="w-12 h-12 rounded-full bg-warning/20 flex items-center justify-center mx-auto mb-4">
            <AlertCircle className="h-6 w-6 text-warning" />
          </div>
          <h2 className="text-lg font-semibold text-text-primary mb-2">
            Download Limit Reached
          </h2>
          <p className="text-text-secondary">
            This file has reached its maximum download limit.
          </p>
        </Card>
      </div>
    );
  }

  // Password protected file
  if (file.password_protected && !downloadStarted) {
    return (
      <div className="container mx-auto px-4 py-20 max-w-md">
        <Card className="py-8 px-6">
          <div className="text-center mb-6">
            <div className="w-12 h-12 rounded-full bg-surface-hover flex items-center justify-center mx-auto mb-4">
              <Lock className="h-6 w-6 text-text-secondary" />
            </div>
            <h2 className="text-lg font-semibold text-text-primary mb-1">
              Password Protected
            </h2>
            <p className="text-sm text-text-secondary">
              Enter the password to download this file
            </p>
          </div>
          
          <form onSubmit={handlePasswordSubmit} className="space-y-4">
            <Input
              type="password"
              placeholder="Enter password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              error={passwordError}
              autoFocus
            />
            <Button
              type="submit"
              className="w-full"
              loading={downloading}
              disabled={!password}
            >
              <DownloadIcon className="h-4 w-4" />
              Download
            </Button>
          </form>
          
          <div className="mt-6 pt-4 border-t border-border text-center">
            <p className="text-sm text-text-secondary">
              {file.filename}
            </p>
            <p className="text-xs text-text-secondary mt-1">
              {formatBytes(file.file_size)}
            </p>
          </div>
        </Card>
      </div>
    );
  }

  // Download started / in progress
  return (
    <div className="container mx-auto px-4 py-20 max-w-md">
      <Card className="text-center py-12">
        {downloading ? (
          <>
            <Spinner size="lg" className="mx-auto" />
            <h2 className="text-lg font-semibold text-text-primary mt-4 mb-2">
              Downloading...
            </h2>
            <p className="text-text-secondary">{file.filename}</p>
          </>
        ) : (
          <>
            <div className="w-12 h-12 rounded-full bg-accent/20 flex items-center justify-center mx-auto mb-4">
              <CheckCircle className="h-6 w-6 text-accent" />
            </div>
            <h2 className="text-lg font-semibold text-text-primary mb-2">
              Download Started
            </h2>
            <p className="text-text-secondary mb-4">
              Your download should begin automatically.
            </p>
            <Button variant="secondary" onClick={() => triggerDownload()}>
              Download Again
            </Button>
          </>
        )}
      </Card>
    </div>
  );
}
