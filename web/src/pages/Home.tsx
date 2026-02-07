import { useState } from 'react';
import { ChevronDown, ChevronUp, Upload } from 'lucide-react';
import { Input, Select, Card, Button } from '@/components/ui';
import { FileUploader } from '@/components/upload/FileUploader';
import { UploadProgress } from '@/components/upload/UploadProgress';
import { UploadSuccess } from '@/components/upload/UploadSuccess';
import { useUploadMultiple, usePublicStatus } from '@/hooks';
import { EXPIRY_OPTIONS, DEFAULT_EXPIRY } from '@/lib/constants';
import type { UploadOptions } from '@/types';
import { cn } from '@/lib/utils';

export default function Home() {
  const { 
    uploadFiles, 
    progress, 
    results, 
    isUploading, 
    isComplete, 
    reset,
    currentFileIndex,
    totalFiles 
  } = useUploadMultiple();
  const { data: status } = usePublicStatus();
  
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [options, setOptions] = useState<UploadOptions>({
    expires_in: DEFAULT_EXPIRY,
  });

  const handleFileDrop = (files: File[]) => {
    setSelectedFiles((prev) => [...prev, ...files]);
  };

  const handleClearFiles = () => {
    setSelectedFiles([]);
  };

  const handleUpload = async () => {
    if (selectedFiles.length > 0) {
      await uploadFiles(selectedFiles, options);
      setSelectedFiles([]);
    }
  };

  const handleUploadAnother = () => {
    reset();
    setSelectedFiles([]);
    setOptions({ expires_in: DEFAULT_EXPIRY });
    setShowAdvanced(false);
  };

  const isServiceReady = status?.connected;

  // Show success state
  if (isComplete && results.length > 0) {
    return (
      <div className="container mx-auto px-4 py-12 max-w-xl">
        <UploadSuccess results={results} onUploadAnother={handleUploadAnother} />
      </div>
    );
  }

  // Show upload progress
  if (isUploading && progress) {
    return (
      <div className="container mx-auto px-4 py-12 max-w-xl">
        <UploadProgress 
          progress={progress} 
          currentFile={currentFileIndex + 1}
          totalFiles={totalFiles}
        />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-12 max-w-xl">
      {/* Hero text */}
      <div className="text-center mb-8">
        <h1 className="text-3xl font-bold text-text-primary mb-2">
          Share Files Securely
        </h1>
        <p className="text-text-secondary">
          Upload files up to 2GB each. Files expire automatically after 30 days.
        </p>
      </div>

      {/* Service status warning */}
      {!isServiceReady && (
        <Card className="mb-6 bg-warning/10 border-warning/20">
          <div className="flex items-center gap-3 text-warning">
            <div className="w-2 h-2 rounded-full bg-warning" />
            <p className="text-sm">
              Service is currently unavailable. Please try again later.
            </p>
          </div>
        </Card>
      )}

      {/* Upload zone */}
      <Card padding="lg" className="mb-4">
        <FileUploader
          onDrop={handleFileDrop}
          disabled={!isServiceReady || isUploading}
          selectedFiles={selectedFiles}
          onClearFiles={handleClearFiles}
        />
        
        {selectedFiles.length > 0 && (
          <div className="mt-4 pt-4 border-t border-border">
            <Button 
              className="w-full" 
              onClick={handleUpload}
              disabled={!isServiceReady}
            >
              <Upload className="h-4 w-4 mr-2" />
              Upload {selectedFiles.length} file{selectedFiles.length !== 1 ? 's' : ''}
            </Button>
          </div>
        )}
      </Card>

      {/* Advanced options */}
      <Card padding="none" className="overflow-hidden">
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="w-full px-4 py-3 flex items-center justify-between text-text-secondary hover:text-text-primary transition-colors"
        >
          <span className="text-sm font-medium">Advanced Options</span>
          {showAdvanced ? (
            <ChevronUp className="h-4 w-4" />
          ) : (
            <ChevronDown className="h-4 w-4" />
          )}
        </button>
        
        <div
          className={cn(
            'px-4 pb-4 space-y-4 transition-all duration-200',
            showAdvanced ? 'block' : 'hidden'
          )}
        >
          <Input
            label="Description (optional)"
            placeholder="Add a description for your files"
            value={options.description || ''}
            onChange={(e) => setOptions({ ...options, description: e.target.value })}
          />
          
          <Input
            label="Password (optional)"
            type="password"
            placeholder="Protect with a password"
            value={options.password || ''}
            onChange={(e) => setOptions({ ...options, password: e.target.value })}
          />
          
          <Select
            label="Expires in"
            options={EXPIRY_OPTIONS.map((o) => ({ label: o.label, value: o.value }))}
            value={options.expires_in || DEFAULT_EXPIRY}
            onChange={(e) => setOptions({ ...options, expires_in: Number(e.target.value) })}
          />
          
          <Input
            label="Max downloads (optional)"
            type="number"
            placeholder="Unlimited"
            min={1}
            value={options.max_downloads || ''}
            onChange={(e) => setOptions({
              ...options,
              max_downloads: e.target.value ? Number(e.target.value) : undefined,
            })}
          />
        </div>
      </Card>
    </div>
  );
}
