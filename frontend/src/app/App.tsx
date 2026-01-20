import { RouterProvider } from 'react-router-dom';
import { QueryProvider } from './providers';
import { router } from './router';

export function App() {
  return (
    <QueryProvider>
      <RouterProvider router={router} />
    </QueryProvider>
  );
}
