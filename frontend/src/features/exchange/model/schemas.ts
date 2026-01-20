import { z } from 'zod';

export const exchangeSchema = z.object({
  sourceAccountId: z.string().uuid('Select a source account'),
  targetAccountId: z.string().uuid('Select a target account'),
  amount: z
    .string()
    .min(1, 'Amount is required')
    .refine((val) => !isNaN(parseFloat(val)) && parseFloat(val) > 0, {
      message: 'Amount must be greater than 0',
    }),
});

export type ExchangeFormData = z.infer<typeof exchangeSchema>;
