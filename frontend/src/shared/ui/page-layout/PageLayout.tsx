import { ReactNode } from 'react';

interface PageLayoutProps {
  title?: string;
  children: ReactNode;
  actions?: ReactNode;
}

export function PageLayout({ title, children, actions }: PageLayoutProps) {
  return (
    <div className="space-y-6">
      {(title || actions) && (
        <div className="flex items-center justify-between">
          {title && <h1 className="text-2xl font-bold text-gray-900">{title}</h1>}
          {actions && <div>{actions}</div>}
        </div>
      )}
      {children}
    </div>
  );
}
