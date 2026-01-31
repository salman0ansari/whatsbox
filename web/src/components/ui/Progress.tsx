import { cn } from '@/lib/utils';

export interface ProgressProps {
  value: number;
  max?: number;
  className?: string;
  showLabel?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export function Progress({ value, max = 100, className, showLabel = false, size = 'md' }: ProgressProps) {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);
  
  const sizes = {
    sm: 'h-1',
    md: 'h-2',
    lg: 'h-3',
  };

  return (
    <div className={cn('w-full', className)}>
      <div className={cn('w-full bg-border rounded-full overflow-hidden', sizes[size])}>
        <div
          className="h-full bg-accent rounded-full transition-all duration-300 ease-out"
          style={{ width: `${percentage}%` }}
        />
      </div>
      {showLabel && (
        <p className="text-xs text-text-secondary mt-1 text-right">
          {Math.round(percentage)}%
        </p>
      )}
    </div>
  );
}
