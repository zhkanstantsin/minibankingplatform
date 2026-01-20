import { Link, useNavigate } from 'react-router-dom';
import { Stack, Button } from '@shared/ui';
import { useCurrentUser } from '@entities/user';
import { useLogout } from '@features/auth';
import { ReconcileIndicator } from '@features/reconcile';

export function Header() {
  const navigate = useNavigate();
  const { data: user } = useCurrentUser();
  const logout = useLogout();

  const handleLogout = () => {
    logout.mutate(undefined, {
      onSuccess: () => navigate('/login'),
    });
  };

  return (
    <header className="border-b border-gray-200 bg-white">
      <div className="container mx-auto px-4">
        <Stack direction="row" justify="between" align="center" className="h-16">
          <Stack direction="row" gap="sm" align="center">
            <Link to="/" className="text-xl font-bold text-gray-900">
              Mini Bank
            </Link>
            <ReconcileIndicator />
          </Stack>

          <nav>
            <Stack direction="row" gap="lg" align="center">
              <Link
                to="/"
                className="text-gray-600 hover:text-gray-900 transition-colors"
              >
                Dashboard
              </Link>
              <Link
                to="/transactions"
                className="text-gray-600 hover:text-gray-900 transition-colors"
              >
                History
              </Link>

              <Stack direction="row" gap="md" align="center">
                <span className="text-sm text-gray-500">{user?.email}</span>
                <Button variant="secondary" size="sm" onClick={handleLogout}>
                  Logout
                </Button>
              </Stack>
            </Stack>
          </nav>
        </Stack>
      </div>
    </header>
  );
}
