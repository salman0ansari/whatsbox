import { useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import { Upload, File } from 'lucide-react';
import { cn, formatBytes } from '@/lib/utils';
import { MAX_FILE_SIZE } from '@/lib/constants';

export interface FileUploaderProps {
  onDrop: (files: File[]) => void;
  disabled?: boolean;
  className?: string;
}

export function FileUploader({ onDrop, disabled, className }: FileUploaderProps) {
  const handleDrop = useCallback(
    (acceptedFiles: File[]) => {
      if (acceptedFiles.length > 0) {
        onDrop(acceptedFiles);
      }
    },
    [onDrop]
  );

  const { getRootProps, getInputProps, isDragActive, isDragReject } = useDropzone({
    onDrop: handleDrop,
    disabled,
    maxSize: MAX_FILE_SIZE,
    multiple: false,
  });

  return (
    <div
      {...getRootProps()}
      className={cn(
        'relative border-2 border-dashed rounded-xl p-8 transition-all duration-200 cursor-pointer',
        'hover:border-accent/50 hover:bg-accent/5',
        isDragActive && 'border-accent bg-accent/10',
        isDragReject && 'border-error bg-error/10',
        disabled && 'opacity-50 cursor-not-allowed hover:border-border hover:bg-transparent',
        !isDragActive && !isDragReject && 'border-border',
        className
      )}
    >
      <input {...getInputProps()} />
      
      <div className="flex flex-col items-center text-center">
        <div
          className={cn(
            'w-14 h-14 rounded-full flex items-center justify-center mb-4 transition-colors',
            isDragActive ? 'bg-accent/20' : 'bg-surface-hover'
          )}
        >
          {isDragActive ? (
            <File className="h-7 w-7 text-accent" />
          ) : (
            <Upload className="h-7 w-7 text-text-secondary" />
          )}
        </div>
        
        <h3 className="text-lg font-medium text-text-primary mb-1">
          {isDragActive ? 'Drop your file here' : 'Drop file here or click to browse'}
        </h3>
        
        <p className="text-sm text-text-secondary">
          Max file size: {formatBytes(MAX_FILE_SIZE)}
        </p>
      </div>
    </div>
  );
}
