import { apiClient } from './client';
import type { FileItem, FileListResponse, UploadOptions, UploadResponse } from '@/types';

export interface ListFilesParams {
  page?: number;
  per_page?: number;
  status?: 'active' | 'expired' | 'deleted';
}

export async function listFiles(params: ListFilesParams = {}): Promise<FileListResponse> {
  const { data } = await apiClient.get<FileListResponse>('/files', { params });
  return data;
}

export async function getFile(id: string): Promise<FileItem> {
  const { data } = await apiClient.get<FileItem>(`/files/${id}`);
  return data;
}

export async function uploadFile(file: File, options: UploadOptions = {}): Promise<UploadResponse> {
  const formData = new FormData();
  formData.append('file', file);
  
  if (options.description) {
    formData.append('description', options.description);
  }
  if (options.password) {
    formData.append('password', options.password);
  }
  if (options.max_downloads !== undefined) {
    formData.append('max_downloads', options.max_downloads.toString());
  }
  if (options.expires_in !== undefined) {
    formData.append('expires_in', options.expires_in.toString());
  }

  const { data } = await apiClient.post<UploadResponse>('/files', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return data;
}

export async function deleteFile(id: string): Promise<void> {
  await apiClient.delete(`/files/${id}`);
}

export async function downloadFile(id: string, password?: string): Promise<Blob> {
  const headers: Record<string, string> = {};
  if (password) {
    headers['X-Password'] = password;
  }

  const { data } = await apiClient.get(`/files/${id}/download`, {
    headers,
    responseType: 'blob',
  });
  return data;
}

export function getDownloadUrl(id: string): string {
  return `/api/files/${id}/download`;
}
