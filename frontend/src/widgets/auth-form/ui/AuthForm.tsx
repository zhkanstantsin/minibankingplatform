import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from 'react-router-dom';
import {
  Card,
  CardHeader,
  CardContent,
  CardFooter,
  Stack,
  FormField,
  Input,
  Button,
  Alert,
} from '@shared/ui';
import { getErrorMessage } from '@shared/api';
import {
  useLogin,
  useRegister,
  loginSchema,
  registerSchema,
  type LoginFormData,
  type RegisterFormData,
} from '@features/auth';

interface AuthFormProps {
  mode: 'login' | 'register';
  onSuccess?: () => void;
}

export function AuthForm({ mode, onSuccess }: AuthFormProps) {
  const login = useLogin();
  const register = useRegister();
  const mutation = mode === 'login' ? login : register;
  const schema = mode === 'login' ? loginSchema : registerSchema;

  const form = useForm<LoginFormData | RegisterFormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const onSubmit = form.handleSubmit((data) => {
    mutation.mutate(data as LoginFormData & RegisterFormData, { onSuccess });
  });

  return (
    <Card className="w-full max-w-md mx-auto">
      <CardHeader>
        <h2 className="text-xl font-semibold text-center">
          {mode === 'login' ? 'Sign In' : 'Create Account'}
        </h2>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit}>
          <Stack direction="column" gap="md">
            <FormField
              label="Email"
              error={form.formState.errors.email?.message}
            >
              <Input
                type="email"
                placeholder="you@example.com"
                error={!!form.formState.errors.email}
                {...form.register('email')}
              />
            </FormField>

            <FormField
              label="Password"
              error={form.formState.errors.password?.message}
            >
              <Input
                type="password"
                placeholder={mode === 'register' ? 'Min 8 characters' : ''}
                error={!!form.formState.errors.password}
                {...form.register('password')}
              />
            </FormField>

            {mutation.isError && (
              <Alert variant="error">{getErrorMessage(mutation.error)}</Alert>
            )}

            <Button type="submit" loading={mutation.isPending} className="w-full">
              {mode === 'login' ? 'Sign In' : 'Create Account'}
            </Button>
          </Stack>
        </form>
      </CardContent>
      <CardFooter className="text-center">
        <Link
          to={mode === 'login' ? '/register' : '/login'}
          className="text-sm text-blue-600 hover:underline"
        >
          {mode === 'login'
            ? "Don't have an account? Sign up"
            : 'Already have an account? Sign in'}
        </Link>
      </CardFooter>
    </Card>
  );
}
