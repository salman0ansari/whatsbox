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

export function useUpload() {
  const queryClient = useQueryClient();
  const [progress, setProgress] = useState<UploadProgress | null>(null);
  const [result, setResult] = useState<UploadResponse | null>(null);
  const uploadRef = useRef<tus.Upload | null>(null);

  const upload = useCallback(async (file: File, options: UploadOptions = {}) => {
    // Validate file size
    if (file.size > MAX_FILE_SIZE) {
      toast.error('File too large. Maximum size is 2GB.');
      return;
    }

    setProgress({
      file,
      progress: 0,
      bytesUploaded: 0,
      bytesTotal: file.size,
      status: 'pending',
    });
    setResult(null);

    try {
      let uploadResult: UploadResponse;

      // Use chunked upload for large files
      if (file.size > CHUNKED_UPLOAD_THRESHOLD) {
        uploadResult = await new Promise<UploadResponse>((resolve, reject) => {
          const tusUpload = createTusUpload(file, options, {
            onProgress: (p) => setProgress(p),
            onSuccess: (res) => {
              // After tus upload, fetch the full file info
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
          startUpload(tusUpload).catch(reject);
        });
      } else {
        // Use simple multipart upload for smaller files
        setProgress((prev) => prev ? { ...prev, status: 'uploading' } : null);
        uploadResult = await uploadFile(file, options);
      }

      setProgress((prev) => prev ? { ...prev, status: 'complete', progress: 100, result: uploadResult } : null);
      setResult(uploadResult);
      toast.success('File uploaded successfully!');
      queryClient.invalidateQueries({ queryKey: ['files'] });
      
      return uploadResult;
    } catch (error) {
      const message = getErrorMessage(error);
      setProgress((prev) => prev ? { ...prev, status: 'error', error: message } : null);
      toast.error(message);
      throw error;
    }
  }, [queryClient]);

  const cancel = useCallback(() => {
    if (uploadRef.current) {
      abortUpload(uploadRef.current);
      uploadRef.current = null;
    }
    setProgress(null);
    setResult(null);
  }, []);

  const reset = useCallback(() => {
    setProgress(null);
    setResult(null);
  }, []);

  return {
    upload,
    cancel,
    reset,
    progress,
    result,
    isUploading: progress?.status === 'uploading' || progress?.status === 'processing',
    isComplete: progress?.status === 'complete',
    isError: progress?.status === 'error',
  };
}
