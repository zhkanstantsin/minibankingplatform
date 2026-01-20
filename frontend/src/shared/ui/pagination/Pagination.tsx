import { Button } from '../button';

interface PaginationProps {
  current: number;
  total: number;
  onChange: (page: number) => void;
}

export function Pagination({ current, total, onChange }: PaginationProps) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-gray-600">
        Page {current} of {total}
      </span>
      <div className="flex gap-2">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onChange(current - 1)}
          disabled={current <= 1}
        >
          Previous
        </Button>
        <Button
          variant="secondary"
          size="sm"
          onClick={() => onChange(current + 1)}
          disabled={current >= total}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
