import { HTMLAttributes } from 'react';
import { cn } from '@shared/lib';

interface AlertProps extends HTMLAttributes<HTMLDivElement> {
  variant?: 'info' | 'success' | 'warning' | 'error';
}

export function Alert({ className, variant = 'info', children, ...props }: AlertProps) {
  const variants = {
    info: 'bg-blue-50 text-blue-800 border-blue-200',
    success: 'bg-green-50 text-green-800 border-green-200',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-200',
    error: 'bg-red-50 text-red-800 border-red-200',
  };

  return (
    <div
      role="alert"
      className={cn(
        'px-4 py-3 rounded-lg border text-sm',
        variants[variant],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
