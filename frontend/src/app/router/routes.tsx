import { createBrowserRouter } from 'react-router-dom';
import { ProtectedLayout, PublicLayout } from '../layouts';
import { LoginPage } from '@pages/login';
import { RegisterPage } from '@pages/register';
import { DashboardPage } from '@pages/dashboard';
import { TransactionsPage } from '@pages/transactions';

export const router = createBrowserRouter([
  {
    element: <PublicLayout />,
    children: [
      {
        path: '/login',
        element: <LoginPage />,
      },
      {
        path: '/register',
        element: <RegisterPage />,
      },
    ],
  },
  {
    element: <ProtectedLayout />,
    children: [
      {
        path: '/',
        element: <DashboardPage />,
      },
      {
        path: '/transactions',
        element: <TransactionsPage />,
      },
    ],
  },
]);
