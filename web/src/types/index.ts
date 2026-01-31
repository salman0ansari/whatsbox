// File types
export interface FileItem {
  id: string;
  filename: string;
  mime_type: string;
  file_size: number;
  description?: string;
  download_url: string;
  password_protected: boolean;
  max_downloads?: number;
  download_count: number;
  created_at: string;
  expires_at: string;
  status: 'active' | 'expired' | 'deleted';
}

export interface FileListResponse {
  files: FileItem[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface UploadOptions {
  description?: string;
  password?: string;
  max_downloads?: number;
  expires_in?: number; // seconds
}

export interface UploadResponse {
  id: string;
  filename: string;
  file_size: number;
  download_url: string;
  expires_at: string;
}

// Admin types
export interface ConnectionStatus {
  connected: boolean;
  logged_in: boolean;
  connected_at?: string;
  reconnect_count: number;
  phone_number?: string;
  push_name?: string;
}

export interface QRResponse {
  qr_code: string; // base64 PNG
}

export interface RealtimeStats {
  uploads_total: number;
  downloads_total: number;
  bytes_uploaded: number;
  bytes_downloaded: number;
  active_uploads: number;
  active_downloads: number;
  upload_errors: number;
  download_errors: number;
  uptime_seconds: number;
  start_time: string;
}

export interface StorageStats {
  total_files: number;
  active_files: number;
  expired_files: number;
  deleted_files: number;
  total_bytes: number;
}

export interface Stats {
  realtime: RealtimeStats;
  storage: StorageStats;
}

export interface HourlyStats {
  hour: string;
  uploads: number;
  downloads: number;
  bytes_uploaded: number;
  bytes_downloaded: number;
}

export interface DailyStats {
  date: string;
  uploads: number;
  downloads: number;
  bytes_uploaded: number;
  bytes_downloaded: number;
}

// Auth types
export interface LoginRequest {
  password: string;
}

export interface AuthResponse {
  success: boolean;
  message?: string;
}

export interface User {
  authenticated: boolean;
}

// API Error
export interface ApiError {
  error: string;
  message: string;
  request_id?: string;
}

// Upload progress
export interface UploadProgress {
  file: File;
  progress: number;
  bytesUploaded: number;
  bytesTotal: number;
  status: 'pending' | 'uploading' | 'processing' | 'complete' | 'error';
  error?: string;
  result?: UploadResponse;
}
