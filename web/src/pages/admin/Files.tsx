import { useState } from 'react';
import { 
  Search, 
  Trash2, 
  Copy, 
  Check,
  File,
  ChevronLeft,
  ChevronRight,
  Lock,
  Files as FilesIcon
} from 'lucide-react';
import { 
  Card, 
  Button, 
  Input, 
  Badge, 
  EmptyState,
  Loading,
  Modal
} from '@/components/ui';
import { useFiles, useDeleteFile } from '@/hooks';
import { 
  formatBytes, 
  formatTimeRemaining,
  getShareUrl,
  copyToClipboard,
  truncateFilename
} from '@/lib/utils';
import { toast } from 'sonner';

export default function AdminFiles() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [fileToDelete, setFileToDelete] = useState<string | null>(null);

  const { data, isLoading } = useFiles({ page, per_page: 10 });
  const deleteMutation = useDeleteFile();

  const handleCopyLink = async (fileId: string) => {
    try {
      await copyToClipboard(getShareUrl(fileId));
      setCopiedId(fileId);
      toast.success('Link copied');
      setTimeout(() => setCopiedId(null), 2000);
    } catch {
      toast.error('Failed to copy link');
    }
  };

  const handleDeleteClick = (fileId: string) => {
    setFileToDelete(fileId);
    setDeleteModalOpen(true);
  };

  const handleDeleteConfirm = () => {
    if (fileToDelete) {
      deleteMutation.mutate(fileToDelete);
      setDeleteModalOpen(false);
      setFileToDelete(null);
    }
  };

  // Filter files by search
  const filteredFiles = data?.files.filter((file) =>
    file.filename.toLowerCase().includes(search.toLowerCase())
  ) || [];

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-full">
        <Loading message="Loading files..." />
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Files</h1>
          <p className="text-text-secondary mt-1">
            Manage your uploaded files
          </p>
        </div>
      </div>

      {/* Search */}
      <div className="mb-6">
        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-secondary" />
          <Input
            placeholder="Search files..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>

      {/* File list */}
      {filteredFiles.length === 0 ? (
        <Card>
          <EmptyState
            icon={FilesIcon}
            title="No files found"
            description={search ? 'Try a different search term' : 'Upload some files to get started'}
          />
        </Card>
      ) : (
        <div className="space-y-3">
          {filteredFiles.map((file) => (
            <Card key={file.id} padding="md">
              <div className="flex items-center gap-4">
                {/* Icon */}
                <div className="w-10 h-10 rounded-lg bg-surface-hover flex items-center justify-center flex-shrink-0">
                  <File className="h-5 w-5 text-text-secondary" />
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-medium text-text-primary truncate">
                      {truncateFilename(file.filename, 40)}
                    </p>
                    {file.password_protected && (
                      <Lock className="h-3.5 w-3.5 text-text-secondary flex-shrink-0" />
                    )}
                  </div>
                  <div className="flex items-center gap-2 mt-1 text-sm text-text-secondary">
                    <span>{formatBytes(file.file_size)}</span>
                    <span className="text-border">•</span>
                    <span>{file.download_count} downloads</span>
                    <span className="text-border">•</span>
                    <span>Expires in {formatTimeRemaining(file.expires_at)}</span>
                  </div>
                </div>

                {/* Status */}
                <Badge
                  variant={file.status === 'active' ? 'success' : file.status === 'expired' ? 'warning' : 'error'}
                >
                  {file.status}
                </Badge>

                {/* Actions */}
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleCopyLink(file.id)}
                    title="Copy link"
                  >
                    {copiedId === file.id ? (
                      <Check className="h-4 w-4 text-accent" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDeleteClick(file.id)}
                    className="text-error hover:text-error"
                    title="Delete file"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Pagination */}
      {data && data.total_pages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-6">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="text-sm text-text-secondary px-4">
            Page {page} of {data.total_pages}
          </span>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => Math.min(data.total_pages, p + 1))}
            disabled={page === data.total_pages}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      )}

      {/* Delete confirmation modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Delete File"
      >
        <p className="text-text-secondary mb-6">
          Are you sure you want to delete this file? This action cannot be undone.
        </p>
        <div className="flex justify-end gap-3">
          <Button variant="secondary" onClick={() => setDeleteModalOpen(false)}>
            Cancel
          </Button>
          <Button
            variant="danger"
            onClick={handleDeleteConfirm}
            loading={deleteMutation.isPending}
          >
            Delete
          </Button>
        </div>
      </Modal>
    </div>
  );
}
