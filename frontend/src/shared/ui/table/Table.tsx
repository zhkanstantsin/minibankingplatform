import { HTMLAttributes, TdHTMLAttributes, ThHTMLAttributes, forwardRef } from 'react';
import { cn } from '@shared/lib';

interface TableProps extends HTMLAttributes<HTMLTableElement> {}

export const Table = forwardRef<HTMLTableElement, TableProps>(
  ({ className, ...props }, ref) => {
    return (
      <div className="overflow-x-auto">
        <table
          ref={ref}
          className={cn('w-full text-sm text-left', className)}
          {...props}
        />
      </div>
    );
  }
);

Table.displayName = 'Table';

interface TableHeadProps extends HTMLAttributes<HTMLTableSectionElement> {}

export const TableHead = forwardRef<HTMLTableSectionElement, TableHeadProps>(
  ({ className, ...props }, ref) => {
    return (
      <thead
        ref={ref}
        className={cn('text-xs text-gray-700 uppercase bg-gray-50', className)}
        {...props}
      />
    );
  }
);

TableHead.displayName = 'TableHead';

interface TableBodyProps extends HTMLAttributes<HTMLTableSectionElement> {}

export const TableBody = forwardRef<HTMLTableSectionElement, TableBodyProps>(
  ({ className, ...props }, ref) => {
    return (
      <tbody
        ref={ref}
        className={cn('divide-y divide-gray-200', className)}
        {...props}
      />
    );
  }
);

TableBody.displayName = 'TableBody';

interface TableRowProps extends HTMLAttributes<HTMLTableRowElement> {}

export const TableRow = forwardRef<HTMLTableRowElement, TableRowProps>(
  ({ className, ...props }, ref) => {
    return (
      <tr
        ref={ref}
        className={cn('hover:bg-gray-50', className)}
        {...props}
      />
    );
  }
);

TableRow.displayName = 'TableRow';

interface TableCellProps extends TdHTMLAttributes<HTMLTableCellElement> {
  header?: boolean;
}

export const TableCell = forwardRef<HTMLTableCellElement, TableCellProps>(
  ({ className, header, ...props }, ref) => {
    const Component = header ? 'th' : 'td';
    return (
      <Component
        ref={ref as never}
        className={cn(
          'px-4 py-3',
          header && 'font-medium',
          className
        )}
        {...(props as ThHTMLAttributes<HTMLTableCellElement>)}
      />
    );
  }
);

TableCell.displayName = 'TableCell';
