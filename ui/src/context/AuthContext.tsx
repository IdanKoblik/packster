import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api, ApiError } from '../api/client';

interface AuthContextType {
  token: string | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (token: string) => Promise<void>;
  logout: () => void;
  error: string | null;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = 'artifactor_token';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY));
  const [isAdmin, setIsAdmin] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const validateStoredToken = async () => {
      const storedToken = localStorage.getItem(TOKEN_KEY);
      if (storedToken) {
        try {
          const response = await api.validateToken(storedToken);
          if (response.valid) {
            setToken(storedToken);
            setIsAdmin(response.admin);
          } else {
            localStorage.removeItem(TOKEN_KEY);
            setToken(null);
            setIsAdmin(false);
          }
        } catch {
          localStorage.removeItem(TOKEN_KEY);
          setToken(null);
          setIsAdmin(false);
        }
      }
      setIsLoading(false);
    };

    validateStoredToken();
  }, []);

  const login = async (newToken: string) => {
    setError(null);
    setIsLoading(true);
    try {
      const response = await api.validateToken(newToken);
      if (!response.valid) {
        throw new Error('Invalid token');
      }
      if (!response.admin) {
        throw new Error('Admin privileges required');
      }
      localStorage.setItem(TOKEN_KEY, newToken);
      setToken(newToken);
      setIsAdmin(true);
    } catch (err) {
      const message = err instanceof ApiError ? err.message : 'Login failed';
      setError(message);
      throw new Error(message);
    } finally {
      setIsLoading(false);
    }
  };

  const logout = () => {
    localStorage.removeItem(TOKEN_KEY);
    setToken(null);
    setIsAdmin(false);
    setError(null);
  };

  return (
    <AuthContext.Provider
      value={{
        token,
        isAuthenticated: !!token,
        isAdmin,
        login,
        logout,
        error,
        isLoading,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
