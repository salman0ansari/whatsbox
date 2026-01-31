export const API_BASE_URL = '/api';

export const MAX_FILE_SIZE = 2 * 1024 * 1024 * 1024; // 2GB

export const EXPIRY_OPTIONS = [
  { label: '1 hour', value: 3600 },
  { label: '6 hours', value: 21600 },
  { label: '1 day', value: 86400 },
  { label: '7 days', value: 604800 },
  { label: '14 days', value: 1209600 },
  { label: '30 days', value: 2592000 },
] as const;

export const DEFAULT_EXPIRY = 2592000; // 30 days

export const SHARE_PLATFORMS = [
  { id: 'twitter', name: 'Twitter/X', icon: 'Twitter' },
  { id: 'facebook', name: 'Facebook', icon: 'Facebook' },
  { id: 'whatsapp', name: 'WhatsApp', icon: 'MessageCircle' },
  { id: 'telegram', name: 'Telegram', icon: 'Send' },
] as const;

export const FILE_ICONS = {
  image: 'Image',
  video: 'Video',
  audio: 'Music',
  pdf: 'FileText',
  archive: 'Archive',
  document: 'FileText',
  spreadsheet: 'Table',
  presentation: 'Presentation',
  file: 'File',
} as const;
