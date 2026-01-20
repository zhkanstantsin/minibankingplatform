import { Navigate, Outlet } from 'react-router-dom';
import { useAuthStore } from '@features/auth';
import { Header } from '@widgets/header';

export function ProtectedLayout() {
  const token = useAuthStore((s) => s.token);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main className="container mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
