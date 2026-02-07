import { useState } from 'react';
import { FileImage, FileVideo, FileAudio, FileText, FileArchive, File as FileIcon, ExternalLink } from 'lucide-react';
import { Modal, Button } from '@/components/ui';
import { cn, formatBytes } from '@/lib/utils';
import type { FileItem } from '@/types';

interface FilePreviewProps {
  file: FileItem;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

function getFileType(filename: string): 'image' | 'video' | 'audio' | 'document' | 'archive' | 'other' {
  const ext = filename.split('.').pop()?.toLowerCase() || '';
  
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'ico'];
  const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv', 'm4v'];
  const audioExts = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a', 'wma'];
  const archiveExts = ['zip', 'rar', '7z', 'tar', 'gz', 'bz2'];
  const docExts = ['pdf', 'doc', 'docx', 'txt', 'rtf', 'odt', 'xls', 'xlsx', 'ppt', 'pptx'];
  
  if (imageExts.includes(ext)) return 'image';
  if (videoExts.includes(ext)) return 'video';
  if (audioExts.includes(ext)) return 'audio';
  if (archiveExts.includes(ext)) return 'archive';
  if (docExts.includes(ext)) return 'document';
  
  return 'other';
}

function getFileIcon(type: string) {
  const iconClass = "h-full w-full";
  
  switch (type) {
    case 'image':
      return <FileImage className={iconClass} />;
    case 'video':
      return <FileVideo className={iconClass} />;
    case 'audio':
      return <FileAudio className={iconClass} />;
    case 'archive':
      return <FileArchive className={iconClass} />;
    case 'document':
      return <FileText className={iconClass} />;
    default:
      return <FileIcon className={iconClass} />;
  }
}

function getFileColor(type: string): string {
  switch (type) {
    case 'image':
      return 'text-purple-400 bg-purple-400/10';
    case 'video':
      return 'text-red-400 bg-red-400/10';
    case 'audio':
      return 'text-yellow-400 bg-yellow-400/10';
    case 'archive':
      return 'text-orange-400 bg-orange-400/10';
    case 'document':
      return 'text-blue-400 bg-blue-400/10';
    default:
      return 'text-text-secondary bg-surface-hover';
  }
}

export function FileTypeIcon({ file, size = 'md', className }: FilePreviewProps) {
  const fileType = getFileType(file.filename);
  const sizeClasses = {
    sm: 'w-8 h-8',
    md: 'w-10 h-10',
    lg: 'w-14 h-14',
  };
  
  const iconSizes = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-7 w-7',
  };
  
  return (
    <div
      className={cn(
        'rounded-lg flex items-center justify-center flex-shrink-0',
        sizeClasses[size],
        getFileColor(fileType),
        className
      )}
    >
      <span className={iconSizes[size]}>
        {getFileIcon(fileType)}
      </span>
    </div>
  );
}

interface FilePreviewModalProps {
  file: FileItem | null;
  isOpen: boolean;
  onClose: () => void;
}

export function FilePreviewModal({ file, isOpen, onClose }: FilePreviewModalProps) {
  const [loading, setLoading] = useState(true);
  
  if (!file) return null;
  
  const fileType = getFileType(file.filename);
  const isImage = fileType === 'image';
  const isVideo = fileType === 'video';
  const isAudio = fileType === 'audio';
  const canPreview = isImage || isVideo || isAudio;
  
  const downloadUrl = `/api/files/${file.id}/download`;
  
  return (
    <Modal isOpen={isOpen} onClose={onClose} title="File Preview" className="max-w-2xl">
      <div className="space-y-4">
        {/* Preview area */}
        <div className="bg-surface-hover rounded-lg overflow-hidden flex items-center justify-center min-h-[200px] max-h-[400px]">
          {canPreview ? (
            <>
              {loading && (
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
                </div>
              )}
              
              {isImage && (
                <img
                  src={downloadUrl}
                  alt={file.filename}
                  className="max-w-full max-h-[400px] object-contain"
                  onLoad={() => setLoading(false)}
                  onError={() => setLoading(false)}
                />
              )}
              
              {isVideo && (
                <video
                  src={downloadUrl}
                  controls
                  className="max-w-full max-h-[400px]"
                  onLoadedData={() => setLoading(false)}
                  onError={() => setLoading(false)}
                />
              )}
              
              {isAudio && (
                <audio
                  src={downloadUrl}
                  controls
                  className="w-full px-4"
                  onLoadedData={() => setLoading(false)}
                  onError={() => setLoading(false)}
                />
              )}
            </>
          ) : (
            <div className="text-center p-8">
              <div className={cn('w-20 h-20 rounded-lg flex items-center justify-center mx-auto mb-4', getFileColor(fileType))}>
                <span className="h-10 w-10">{getFileIcon(fileType)}</span>
              </div>
              <p className="text-text-secondary">Preview not available for this file type</p>
            </div>
          )}
        </div>
        
        {/* File info */}
        <div className="bg-surface rounded-lg p-4">
          <h3 className="font-medium text-text-primary truncate">{file.filename}</h3>
          <p className="text-sm text-text-secondary mt-1">
            {formatBytes(file.file_size)} • {file.mime_type}
          </p>
        </div>
        
        {/* Actions */}
        <div className="flex gap-3">
          <Button
            variant="primary"
            className="flex-1"
            onClick={() => window.open(downloadUrl, '_blank')}
          >
            <ExternalLink className="h-4 w-4 mr-2" />
            Download
          </Button>
        </div>
      </div>
    </Modal>
  );
}

export { getFileType, getFileIcon, getFileColor };
