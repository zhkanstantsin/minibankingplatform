import { z } from 'zod';

export const transferSchema = z.object({
  fromAccountId: z.string().uuid('Select an account'),
  toAccountId: z.string().uuid('Enter a valid recipient account ID'),
  amount: z
    .string()
    .min(1, 'Amount is required')
    .refine((val) => !isNaN(parseFloat(val)) && parseFloat(val) > 0, {
      message: 'Amount must be greater than 0',
    }),
  currency: z.enum(['USD', 'EUR'], { message: 'Select a currency' }),
});

export type TransferFormData = z.infer<typeof transferSchema>;
