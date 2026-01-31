import { File, Loader2 } from 'lucide-react';
import { Card, Progress } from '@/components/ui';
import { formatBytes, truncateFilename } from '@/lib/utils';
import type { UploadProgress as UploadProgressType } from '@/types';

export interface UploadProgressProps {
  progress: UploadProgressType;
}

export function UploadProgress({ progress }: UploadProgressProps) {
  const { file, progress: percent, bytesUploaded, bytesTotal, status } = progress;
  
  const speed = bytesUploaded > 0 ? bytesUploaded / ((Date.now() - performance.now()) / 1000) : 0;

  return (
    <Card padding="lg">
      <div className="flex items-start gap-4">
        <div className="w-12 h-12 rounded-lg bg-surface-hover flex items-center justify-center flex-shrink-0">
          <File className="h-6 w-6 text-text-secondary" />
        </div>
        
        <div className="flex-1 min-w-0">
          <h3 className="font-medium text-text-primary truncate">
            {truncateFilename(file.name, 40)}
          </h3>
          
          <div className="mt-3">
            <Progress value={percent} size="md" />
          </div>
          
          <div className="mt-2 flex items-center justify-between text-sm text-text-secondary">
            <div className="flex items-center gap-2">
              {status === 'processing' ? (
                <>
                  <Loader2 className="h-3 w-3 animate-spin" />
                  <span>Processing...</span>
                </>
              ) : (
                <>
                  <span>{formatBytes(bytesUploaded)} / {formatBytes(bytesTotal)}</span>
                  {speed > 0 && (
                    <>
                      <span className="text-border">•</span>
                      <span>{formatBytes(speed)}/s</span>
                    </>
                  )}
                </>
              )}
            </div>
            
            <span className="font-medium text-accent">
              {Math.round(percent)}%
            </span>
          </div>
        </div>
      </div>
    </Card>
  );
}
