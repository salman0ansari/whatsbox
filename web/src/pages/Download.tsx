import { useState, useEffect, useCallback } from 'react';
import { useParams } from '@tanstack/react-router';
import { Download as DownloadIcon, Lock, AlertCircle, CheckCircle, Copy, Check, Share2, ExternalLink } from 'lucide-react';
import { Card, Button, Input, Spinner, Badge } from '@/components/ui';
import { ShareQRCode } from '@/components/share/ShareQRCode';
import { ShareButtons } from '@/components/share/ShareButtons';
import { useFile } from '@/hooks';
import { downloadFile, getDownloadUrl } from '@/api/files';
import { getErrorMessage } from '@/api/client';
import { formatBytes, getShareUrl, copyToClipboard, truncateFilename } from '@/lib/utils';
import { downloadRoute } from '@/routes';
import { toast } from 'sonner';
import { FileTypeIcon } from '@/components/files/FilePreview';

export default function Download() {
  const { fileId } = useParams({ from: downloadRoute.id });
  const { data: file, isLoading, error } = useFile(fileId);
  
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [downloading, setDownloading] = useState(false);
  const [downloadStarted, setDownloadStarted] = useState(false);
  const [copied, setCopied] = useState(false);
  const [showQR, setShowQR] = useState(false);

  const shareUrl = file ? getShareUrl(file.id) : '';

  const handleCopyLink = async () => {
    try {
      await copyToClipboard(shareUrl);
      setCopied(true);
      toast.success('Link copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy link');
    }
  };

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
      <div className="container mx-auto px-4 py-12 max-w-lg">
        <Card className="py-8 px-6">
          <div className="text-center mb-6">
            <FileTypeIcon file={file} size="lg" className="mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-text-primary mb-1">
              {truncateFilename(file.filename, 50)}
            </h2>
            <div className="flex items-center justify-center gap-2 mt-2">
              <Badge variant="warning" size="sm">
                <Lock className="h-3 w-3 mr-1" />
                Password Protected
              </Badge>
              <span className="text-sm text-text-secondary">
                {formatBytes(file.file_size)}
              </span>
            </div>
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
              <DownloadIcon className="h-4 w-4 mr-2" />
              Download
            </Button>
          </form>
          
          {/* Share Link */}
          <div className="mt-6 pt-4 border-t border-border">
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Share Link
            </label>
            <div className="flex gap-2">
              <div className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-text-primary text-sm truncate">
                {shareUrl}
              </div>
              <Button variant="secondary" size="md" onClick={handleCopyLink}>
                {copied ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </Card>
      </div>
    );
  }

  // Download started / in progress
  return (
    <div className="container mx-auto px-4 py-12 max-w-lg">
      <Card className="py-8 px-6">
        {downloading ? (
          <div className="text-center">
            <Spinner size="lg" className="mx-auto" />
            <h2 className="text-xl font-semibold text-text-primary mt-4 mb-2">
              Downloading...
            </h2>
            <p className="text-text-secondary">{file.filename}</p>
          </div>
        ) : (
          <>
            {/* File Info */}
            <div className="text-center mb-6">
              <FileTypeIcon file={file} size="lg" className="mx-auto mb-4" />
              <h2 className="text-xl font-semibold text-text-primary mb-1">
                {truncateFilename(file.filename, 50)}
              </h2>
              <p className="text-sm text-text-secondary">
                {formatBytes(file.file_size)}
              </p>
            </div>

            {/* Success Message */}
            <div className="flex items-center justify-center gap-2 mb-6 text-accent">
              <CheckCircle className="h-5 w-5" />
              <span className="font-medium">Download Started</span>
            </div>

            {/* Share Link */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Share Link
              </label>
              <div className="flex gap-2">
                <div className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-text-primary text-sm truncate">
                  {shareUrl}
                </div>
                <Button variant="secondary" size="md" onClick={handleCopyLink}>
                  {copied ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>

            {/* QR Code Toggle */}
            <div className="mb-6">
              <button
                onClick={() => setShowQR(!showQR)}
                className="flex items-center justify-center gap-2 w-full py-2 text-sm text-text-secondary hover:text-text-primary transition-colors"
              >
                <Share2 className="h-4 w-4" />
                {showQR ? 'Hide QR Code' : 'Show QR Code'}
              </button>
              
              {showQR && (
                <div className="mt-4 flex justify-center animate-in fade-in duration-200">
                  <ShareQRCode url={shareUrl} size={180} />
                </div>
              )}
            </div>

            {/* Share Buttons */}
            <div className="mb-6 pt-4 border-t border-border">
              <p className="text-sm text-text-secondary mb-3 text-center">Share via</p>
              <ShareButtons url={shareUrl} title={`Download ${file.filename}`} />
            </div>

            {/* Actions */}
            <div className="flex gap-3 pt-4 border-t border-border">
              <Button variant="secondary" className="flex-1" onClick={() => triggerDownload()}>
                <DownloadIcon className="h-4 w-4 mr-2" />
                Download Again
              </Button>
              <Button 
                variant="primary" 
                className="flex-1"
                onClick={() => window.open(getDownloadUrl(file.id), '_blank')}
              >
                <ExternalLink className="h-4 w-4 mr-2" />
                Open File
              </Button>
            </div>
          </>
        )}
      </Card>
    </div>
  );
}
