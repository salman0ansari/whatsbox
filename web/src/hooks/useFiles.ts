import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { listFiles, getFile, deleteFile, type ListFilesParams } from '@/api/files';
import { getErrorMessage } from '@/api/client';

export function useFiles(params: ListFilesParams = {}) {
  return useQuery({
    queryKey: ['files', params],
    queryFn: () => listFiles(params),
  });
}

export function useFile(id: string) {
  return useQuery({
    queryKey: ['file', id],
    queryFn: () => getFile(id),
    enabled: !!id,
  });
}

export function useDeleteFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteFile,
    onSuccess: () => {
      toast.success('File deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['files'] });
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
}
