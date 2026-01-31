import * as tus from 'tus-js-client';
import type { UploadOptions, UploadProgress } from '@/types';

export interface TusUploadCallbacks {
  onProgress?: (progress: UploadProgress) => void;
  onSuccess?: (result: { id: string; filename: string }) => void;
  onError?: (error: Error) => void;
}

export function createTusUpload(
  file: File,
  options: UploadOptions = {},
  callbacks: TusUploadCallbacks = {}
): tus.Upload {
  // Build metadata
  const metadata: Record<string, string> = {
    filename: file.name,
    filetype: file.type || 'application/octet-stream',
  };

  if (options.description) {
    metadata.description = options.description;
  }
  if (options.password) {
    metadata.password = options.password;
  }
  if (options.max_downloads !== undefined) {
    metadata.max_downloads = options.max_downloads.toString();
  }
  if (options.expires_in !== undefined) {
    metadata.expires_in = options.expires_in.toString();
  }

  const upload = new tus.Upload(file, {
    endpoint: '/api/upload',
    retryDelays: [0, 1000, 3000, 5000],
    chunkSize: 5 * 1024 * 1024, // 5MB chunks
    metadata,
    onProgress: (bytesUploaded, bytesTotal) => {
      const progress = (bytesUploaded / bytesTotal) * 100;
      callbacks.onProgress?.({
        file,
        progress,
        bytesUploaded,
        bytesTotal,
        status: progress < 100 ? 'uploading' : 'processing',
      });
    },
    onSuccess: (_payload) => {
      // Extract file ID from upload URL
      const url = upload.url;
      const id = url?.split('/').pop() || '';
      callbacks.onSuccess?.({ id, filename: file.name });
    },
    onError: (error) => {
      callbacks.onError?.(error);
    },
  });

  return upload;
}

export function startUpload(upload: tus.Upload): Promise<void> {
  return new Promise((resolve, reject) => {
    const originalOnSuccess = upload.options.onSuccess;
    const originalOnError = upload.options.onError;
    
    upload.options.onSuccess = (payload) => {
      if (originalOnSuccess) {
        originalOnSuccess(payload);
      }
      resolve();
    };
    upload.options.onError = (error) => {
      if (originalOnError) {
        originalOnError(error);
      }
      reject(error);
    };
    
    // Check for previous uploads
    upload.findPreviousUploads().then((previousUploads) => {
      if (previousUploads.length > 0) {
        upload.resumeFromPreviousUpload(previousUploads[0]);
      }
      upload.start();
    });
  });
}

export function abortUpload(upload: tus.Upload): void {
  upload.abort();
}
