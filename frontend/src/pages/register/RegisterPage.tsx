import { useNavigate } from 'react-router-dom';
import { PageLayout, Stack } from '@shared/ui';
import { AuthForm } from '@widgets/auth-form';

export function RegisterPage() {
  const navigate = useNavigate();

  return (
    <PageLayout>
      <Stack direction="column" align="center" justify="center" className="min-h-[80vh]">
        <AuthForm mode="register" onSuccess={() => navigate('/')} />
      </Stack>
    </PageLayout>
  );
}
