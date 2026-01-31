import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export function formatDate(date: string | Date): string {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(new Date(date));
}

export function formatDateTime(date: string | Date): string {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(new Date(date));
}

export function formatRelativeTime(date: string | Date): string {
  const now = new Date();
  const then = new Date(date);
  const diffInSeconds = Math.floor((now.getTime() - then.getTime()) / 1000);

  if (diffInSeconds < 60) {
    return 'just now';
  }

  const diffInMinutes = Math.floor(diffInSeconds / 60);
  if (diffInMinutes < 60) {
    return `${diffInMinutes}m ago`;
  }

  const diffInHours = Math.floor(diffInMinutes / 60);
  if (diffInHours < 24) {
    return `${diffInHours}h ago`;
  }

  const diffInDays = Math.floor(diffInHours / 24);
  if (diffInDays < 30) {
    return `${diffInDays}d ago`;
  }

  return formatDate(date);
}

export function formatTimeRemaining(expiresAt: string | Date): string {
  const now = new Date();
  const expires = new Date(expiresAt);
  const diffInSeconds = Math.floor((expires.getTime() - now.getTime()) / 1000);

  if (diffInSeconds <= 0) {
    return 'Expired';
  }

  const days = Math.floor(diffInSeconds / (24 * 60 * 60));
  const hours = Math.floor((diffInSeconds % (24 * 60 * 60)) / (60 * 60));

  if (days > 0) {
    return `${days}d ${hours}h`;
  }

  const minutes = Math.floor((diffInSeconds % (60 * 60)) / 60);
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }

  return `${minutes}m`;
}

export function getFileIcon(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'image';
  if (mimeType.startsWith('video/')) return 'video';
  if (mimeType.startsWith('audio/')) return 'audio';
  if (mimeType.includes('pdf')) return 'pdf';
  if (mimeType.includes('zip') || mimeType.includes('rar') || mimeType.includes('tar') || mimeType.includes('gzip')) return 'archive';
  if (mimeType.includes('text') || mimeType.includes('document') || mimeType.includes('word')) return 'document';
  if (mimeType.includes('spreadsheet') || mimeType.includes('excel')) return 'spreadsheet';
  if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return 'presentation';
  return 'file';
}

export function truncateFilename(filename: string, maxLength = 30): string {
  if (filename.length <= maxLength) return filename;
  
  const extension = filename.split('.').pop() || '';
  const nameWithoutExt = filename.slice(0, filename.length - extension.length - 1);
  const truncatedName = nameWithoutExt.slice(0, maxLength - extension.length - 4);
  
  return `${truncatedName}...${extension}`;
}

export function getShareUrl(fileId: string): string {
  const baseUrl = window.location.origin;
  return `${baseUrl}/f/${fileId}`;
}

export function copyToClipboard(text: string): Promise<void> {
  return navigator.clipboard.writeText(text);
}
