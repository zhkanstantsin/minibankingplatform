export interface User {
  userId: string;
  email: string;
}

export interface AuthResponse {
  userId: string;
  email: string;
  token: string;
}
