import { useState, useCallback, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import * as tus from 'tus-js-client';
import { uploadFile } from '@/api/files';
import { createTusUpload, startUpload, abortUpload } from '@/api/upload';
import { getErrorMessage } from '@/api/client';
import type { UploadOptions, UploadProgress, UploadResponse } from '@/types';
import { MAX_FILE_SIZE } from '@/lib/constants';

// Threshold for using chunked uploads (50MB)
const CHUNKED_UPLOAD_THRESHOLD = 50 * 1024 * 1024;

export function useUploadMultiple() {
  const queryClient = useQueryClient();
  const [progress, setProgress] = useState<UploadProgress | null>(null);
  const [results, setResults] = useState<UploadResponse[]>([]);
  const [currentFileIndex, setCurrentFileIndex] = useState(0);
  const [totalFiles, setTotalFiles] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const uploadRef = useRef<tus.Upload | null>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  const uploadSingleFile = useCallback(async (
    file: File, 
    options: UploadOptions
  ): Promise<UploadResponse> => {
    // Validate file size
    if (file.size > MAX_FILE_SIZE) {
      throw new Error(`File "${file.name}" is too large. Maximum size is 2GB.`);
    }

    setProgress({
      file,
      progress: 0,
      bytesUploaded: 0,
      bytesTotal: file.size,
      status: 'pending',
    });

    try {
      let uploadResult: UploadResponse;

      // Use chunked upload for large files
      if (file.size > CHUNKED_UPLOAD_THRESHOLD) {
        uploadResult = await new Promise<UploadResponse>((resolve, reject) => {
          abortControllerRef.current = new AbortController();
          
          const tusUpload = createTusUpload(file, options, {
            onProgress: (p) => setProgress(p),
            onSuccess: (res) => {
              resolve({
                id: res.id,
                filename: res.filename,
                file_size: file.size,
                download_url: `/api/files/${res.id}/download`,
                expires_at: new Date(Date.now() + (options.expires_in || 30 * 24 * 60 * 60) * 1000).toISOString(),
              });
            },
            onError: (error) => reject(error),
          });

          uploadRef.current = tusUpload;
          
          // Handle abort
          abortControllerRef.current.signal.addEventListener('abort', () => {
            abortUpload(tusUpload);
          });
          
          startUpload(tusUpload).catch(reject);
        });
      } else {
        // Use simple multipart upload for smaller files
        abortControllerRef.current = new AbortController();
        
        setProgress((prev) => prev ? { ...prev, status: 'uploading' } : null);
        uploadResult = await uploadFile(file, options);
      }

      setProgress((prev) => prev ? { ...prev, status: 'complete', progress: 100, result: uploadResult } : null);
      
      toast.success(`"${file.name}" uploaded successfully!`);
      
      return uploadResult;
    } catch (error) {
      const message = getErrorMessage(error);
      setProgress((prev) => prev ? { ...prev, status: 'error', error: message } : null);
      throw error;
    }
  }, []);

  const uploadFiles = useCallback(async (files: File[], options: UploadOptions = {}) => {
    setIsUploading(true);
    setResults([]);
    setTotalFiles(files.length);
    setCurrentFileIndex(0);
    
    const uploadedResults: UploadResponse[] = [];
    
    try {
      for (let i = 0; i < files.length; i++) {
        setCurrentFileIndex(i);
        const file = files[i];
        
        try {
          const result = await uploadSingleFile(file, options);
          uploadedResults.push(result);
        } catch (error) {
          const message = getErrorMessage(error);
          toast.error(`Failed to upload "${file.name}": ${message}`);
          // Continue with next file instead of stopping
        }
      }
      
      setResults(uploadedResults);
      
      if (uploadedResults.length > 0) {
        toast.success(`${uploadedResults.length} file${uploadedResults.length !== 1 ? 's' : ''} uploaded successfully!`);
        queryClient.invalidateQueries({ queryKey: ['files'] });
      }
      
      return uploadedResults;
    } catch (error) {
      toast.error('Upload process failed');
      throw error;
    } finally {
      setIsUploading(false);
      abortControllerRef.current = null;
      uploadRef.current = null;
    }
  }, [uploadSingleFile, queryClient]);

  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    if (uploadRef.current) {
      abortUpload(uploadRef.current);
    }
    setProgress(null);
    setIsUploading(false);
  }, []);

  const reset = useCallback(() => {
    setProgress(null);
    setResults([]);
    setCurrentFileIndex(0);
    setTotalFiles(0);
    setIsUploading(false);
    abortControllerRef.current = null;
    uploadRef.current = null;
  }, []);

  return {
    uploadFiles,
    cancel,
    reset,
    progress,
    results,
    currentFileIndex,
    totalFiles,
    isUploading,
    isComplete: !isUploading && results.length > 0 && currentFileIndex >= totalFiles - 1,
    isError: progress?.status === 'error',
  };
}
