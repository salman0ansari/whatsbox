import { useState } from 'react';
import { CheckCircle, Copy, Check, Upload } from 'lucide-react';
import { Card, Button } from '@/components/ui';
import { ShareButtons } from '@/components/share/ShareButtons';
import { ShareQRCode } from '@/components/share/ShareQRCode';
import { formatBytes, getShareUrl, copyToClipboard, formatTimeRemaining } from '@/lib/utils';
import type { UploadResponse } from '@/types';
import { toast } from 'sonner';

export interface UploadSuccessProps {
  result: UploadResponse;
  onUploadAnother: () => void;
}

export function UploadSuccess({ result, onUploadAnother }: UploadSuccessProps) {
  const [copied, setCopied] = useState(false);
  const shareUrl = getShareUrl(result.id);

  const handleCopy = async () => {
    try {
      await copyToClipboard(shareUrl);
      setCopied(true);
      toast.success('Link copied to clipboard');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  return (
    <Card padding="lg">
      <div className="text-center mb-6">
        <div className="w-14 h-14 rounded-full bg-accent/20 flex items-center justify-center mx-auto mb-4">
          <CheckCircle className="h-7 w-7 text-accent" />
        </div>
        <h2 className="text-xl font-semibold text-text-primary mb-1">
          Upload Complete!
        </h2>
        <p className="text-text-secondary">
          Your file is ready to share
        </p>
      </div>

      {/* File info */}
      <div className="bg-surface-hover rounded-lg p-4 mb-6">
        <p className="font-medium text-text-primary truncate">
          {result.filename}
        </p>
        <p className="text-sm text-text-secondary mt-1">
          {formatBytes(result.file_size)} • Expires in {formatTimeRemaining(result.expires_at)}
        </p>
      </div>

      {/* Share link */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-text-secondary mb-2">
          Share Link
        </label>
        <div className="flex gap-2">
          <div className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-text-primary text-sm truncate">
            {shareUrl}
          </div>
          <Button variant="secondary" size="md" onClick={handleCopy}>
            {copied ? (
              <Check className="h-4 w-4 text-accent" />
            ) : (
              <Copy className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>

      {/* QR Code */}
      <div className="flex justify-center mb-6">
        <ShareQRCode url={shareUrl} size={160} />
      </div>

      {/* Share buttons */}
      <div className="mb-6">
        <ShareButtons url={shareUrl} title={`Download ${result.filename}`} />
      </div>

      {/* Upload another */}
      <Button variant="secondary" className="w-full" onClick={onUploadAnother}>
        <Upload className="h-4 w-4" />
        Upload Another File
      </Button>
    </Card>
  );
}
