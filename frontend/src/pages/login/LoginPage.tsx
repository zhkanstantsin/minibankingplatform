import { useNavigate } from 'react-router-dom';
import { PageLayout, Stack } from '@shared/ui';
import { AuthForm } from '@widgets/auth-form';

export function LoginPage() {
  const navigate = useNavigate();

  return (
    <PageLayout>
      <Stack direction="column" align="center" justify="center" className="min-h-[80vh]">
        <AuthForm mode="login" onSuccess={() => navigate('/')} />
      </Stack>
    </PageLayout>
  );
}
