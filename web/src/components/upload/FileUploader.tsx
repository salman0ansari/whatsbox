import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { Upload, File, X, FileImage, FileVideo, FileAudio, FileText, FileArchive } from 'lucide-react';
import { cn, formatBytes } from '@/lib/utils';
import { MAX_FILE_SIZE } from '@/lib/constants';
import { Button } from '@/components/ui';

export interface FileUploaderProps {
  onDrop: (files: File[]) => void;
  disabled?: boolean;
  className?: string;
  selectedFiles?: File[];
  onClearFiles?: () => void;
}

function getFileIcon(filename: string) {
  const ext = filename.split('.').pop()?.toLowerCase();
  
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp'];
  const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv'];
  const audioExts = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a', 'wma'];
  const archiveExts = ['zip', 'rar', '7z', 'tar', 'gz', 'bz2'];
  const docExts = ['pdf', 'doc', 'docx', 'txt', 'rtf', 'odt'];
  
  if (imageExts.includes(ext || '')) return <FileImage className="h-5 w-5" />;
  if (videoExts.includes(ext || '')) return <FileVideo className="h-5 w-5" />;
  if (audioExts.includes(ext || '')) return <FileAudio className="h-5 w-5" />;
  if (archiveExts.includes(ext || '')) return <FileArchive className="h-5 w-5" />;
  if (docExts.includes(ext || '')) return <FileText className="h-5 w-5" />;
  
  return <File className="h-5 w-5" />;
}

export function FileUploader({ 
  onDrop, 
  disabled, 
  className,
  selectedFiles = [],
  onClearFiles 
}: FileUploaderProps) {
  const [files, setFiles] = useState<File[]>(selectedFiles);
  const [errors, setErrors] = useState<string[]>([]);

  const handleDrop = useCallback(
    (acceptedFiles: File[], fileRejections: any[]) => {
      const newErrors: string[] = [];
      
      // Check for rejected files
      fileRejections.forEach((rejection) => {
        rejection.errors.forEach((err: any) => {
          if (err.code === 'file-too-large') {
            newErrors.push(`${rejection.file.name}: File too large (max ${formatBytes(MAX_FILE_SIZE)})`);
          } else {
            newErrors.push(`${rejection.file.name}: ${err.message}`);
          }
        });
      });
      
      if (newErrors.length > 0) {
        setErrors(newErrors);
        setTimeout(() => setErrors([]), 5000);
      }
      
      if (acceptedFiles.length > 0) {
        const newFiles = [...files, ...acceptedFiles];
        setFiles(newFiles);
        onDrop(acceptedFiles);
      }
    },
    [files, onDrop]
  );

  const handleClear = () => {
    setFiles([]);
    setErrors([]);
    onClearFiles?.();
  };

  const removeFile = (index: number) => {
    const newFiles = files.filter((_, i) => i !== index);
    setFiles(newFiles);
  };

  const totalSize = files.reduce((acc, file) => acc + file.size, 0);

  const { getRootProps, getInputProps, isDragActive, isDragReject, open } = useDropzone({
    onDrop: handleDrop,
    disabled,
    maxSize: MAX_FILE_SIZE,
    multiple: true,
    noClick: files.length > 0,
  });

  return (
    <div className={className}>
      <div
        {...getRootProps()}
        className={cn(
          'relative border-2 border-dashed rounded-xl p-8 transition-all duration-200',
          files.length === 0 && 'cursor-pointer',
          'hover:border-accent/50 hover:bg-accent/5',
          isDragActive && 'border-accent bg-accent/10',
          isDragReject && 'border-error bg-error/10',
          disabled && 'opacity-50 cursor-not-allowed hover:border-border hover:bg-transparent',
          !isDragActive && !isDragReject && !disabled && 'border-border',
          files.length > 0 && 'bg-surface-hover/30'
        )}
      >
        <input {...getInputProps()} />
        
        {files.length === 0 ? (
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
              {isDragActive ? 'Drop your files here' : 'Drop files here or click to browse'}
            </h3>
            
            <p className="text-sm text-text-secondary">
              Max file size: {formatBytes(MAX_FILE_SIZE)} • Multiple files supported
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-medium text-text-primary">
                {files.length} file{files.length !== 1 ? 's' : ''} selected
              </h3>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="secondary"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    open();
                  }}
                  disabled={disabled}
                >
                  <Upload className="h-4 w-4 mr-1" />
                  Add more
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleClear();
                  }}
                >
                  <X className="h-4 w-4 mr-1" />
                  Clear
                </Button>
              </div>
            </div>
            
            <div className="max-h-48 overflow-y-auto space-y-2">
              {files.map((file, index) => (
                <div
                  key={`${file.name}-${index}`}
                  className="flex items-center gap-3 p-3 bg-surface rounded-lg group"
                >
                  <div className="text-text-secondary">
                    {getFileIcon(file.name)}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-text-primary truncate">
                      {file.name}
                    </p>
                    <p className="text-xs text-text-secondary">
                      {formatBytes(file.size)}
                    </p>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      removeFile(index);
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 hover:bg-error/10 rounded transition-all"
                    disabled={disabled}
                  >
                    <X className="h-4 w-4 text-text-secondary hover:text-error" />
                  </button>
                </div>
              ))}
            </div>
            
            <div className="flex items-center justify-between pt-2 border-t border-border">
              <p className="text-sm text-text-secondary">
                Total size: <span className="font-medium text-text-primary">{formatBytes(totalSize)}</span>
              </p>
            </div>
          </div>
        )}
      </div>
      
      {errors.length > 0 && (
        <div className="mt-3 space-y-1">
          {errors.map((error, index) => (
            <p key={index} className="text-sm text-error">
              {error}
            </p>
          ))}
        </div>
      )}
    </div>
  );
}
