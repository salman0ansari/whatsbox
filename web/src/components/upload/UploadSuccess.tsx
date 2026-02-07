import { useState } from 'react';
import { CheckCircle, Copy, Check, Upload, File, ChevronDown, ChevronUp } from 'lucide-react';
import { Card, Button } from '@/components/ui';
import { ShareButtons } from '@/components/share/ShareButtons';
import { ShareQRCode } from '@/components/share/ShareQRCode';
import { formatBytes, getShareUrl, copyToClipboard, formatTimeRemaining } from '@/lib/utils';
import type { UploadResponse } from '@/types';
import { toast } from 'sonner';

interface FileResultProps {
  result: UploadResponse;
  isExpanded: boolean;
  onToggle: () => void;
}

function FileResult({ result, isExpanded, onToggle }: FileResultProps) {
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
    <div className="bg-surface-hover rounded-lg overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full p-4 flex items-center gap-3 text-left hover:bg-surface-hover/80 transition-colors"
      >
        <File className="h-5 w-5 text-text-secondary flex-shrink-0" />
        <div className="flex-1 min-w-0">
          <p className="font-medium text-text-primary truncate">
            {result.filename}
          </p>
          <p className="text-sm text-text-secondary">
            {formatBytes(result.file_size)} • Expires in {formatTimeRemaining(result.expires_at)}
          </p>
        </div>
        <Button variant="secondary" size="sm" onClick={(e) => { e.stopPropagation(); handleCopy(); }}>
          {copied ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
        </Button>
        {isExpanded ? <ChevronUp className="h-4 w-4 text-text-secondary" /> : <ChevronDown className="h-4 w-4 text-text-secondary" />}
      </button>
      
      {isExpanded && (
        <div className="px-4 pb-4 border-t border-border/50">
          <div className="mt-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Share Link
              </label>
              <div className="flex gap-2">
                <div className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-text-primary text-sm truncate">
                  {shareUrl}
                </div>
                <Button variant="secondary" size="md" onClick={handleCopy}>
                  {copied ? <Check className="h-4 w-4 text-accent" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>
            
            <div className="flex justify-center">
              <ShareQRCode url={shareUrl} size={140} />
            </div>
            
            <ShareButtons url={shareUrl} title={`Download ${result.filename}`} />
          </div>
        </div>
      )}
    </div>
  );
}

export interface UploadSuccessProps {
  results: UploadResponse | UploadResponse[];
  onUploadAnother: () => void;
}

export function UploadSuccess({ results, onUploadAnother }: UploadSuccessProps) {
  const resultsArray = Array.isArray(results) ? results : [results];
  const [expandedIndex, setExpandedIndex] = useState(0);
  const isMultiple = resultsArray.length > 1;

  return (
    <Card padding="lg">
      <div className="text-center mb-6">
        <div className="w-14 h-14 rounded-full bg-accent/20 flex items-center justify-center mx-auto mb-4">
          <CheckCircle className="h-7 w-7 text-accent" />
        </div>
        <h2 className="text-xl font-semibold text-text-primary mb-1">
          {isMultiple ? `${resultsArray.length} Files Uploaded!` : 'Upload Complete!'}
        </h2>
        <p className="text-text-secondary">
          {isMultiple ? 'Your files are ready to share' : 'Your file is ready to share'}
        </p>
      </div>

      {/* Files list */}
      <div className="space-y-3 mb-6 max-h-96 overflow-y-auto">
        {resultsArray.map((result, index) => (
          <FileResult
            key={result.id}
            result={result}
            isExpanded={expandedIndex === index}
            onToggle={() => setExpandedIndex(expandedIndex === index ? -1 : index)}
          />
        ))}
      </div>

      {/* Upload another */}
      <Button variant="secondary" className="w-full" onClick={onUploadAnother}>
        <Upload className="h-4 w-4 mr-2" />
        Upload More Files
      </Button>
    </Card>
  );
}
