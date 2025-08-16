import React, { createContext, useContext, useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import * as Types from '../types';
import { authApi } from '../lib/api';

interface AuthContextType {
  user: Types.User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
  login: (data: Types.LoginRequest) => Promise<void>;
  register: (data: Types.RegisterRequest) => Promise<void>;
  logout: () => void;
  clearError: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<Types.User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 初期化時にローカルストレージからトークンを復元
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        const storedToken = localStorage.getItem('token');
        if (storedToken) {
          setToken(storedToken);
          // トークンが有効かチェックするためにプロフィール取得
          const userData = await authApi.getProfile();
          setUser(userData);
        }
      } catch (error) {
        // トークンが無効な場合、削除する
        localStorage.removeItem('token');
        setToken(null);
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    initializeAuth();
  }, []);

  // ログイン関数
  const login = async (data: Types.LoginRequest): Promise<void> => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await authApi.login(data);
      
      // トークンとユーザー情報を保存
      localStorage.setItem('token', response.token);
      setToken(response.token);
      setUser(response.user);
      
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'ログインに失敗しました';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // ユーザー登録関数
  const register = async (data: Types.RegisterRequest): Promise<void> => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await authApi.register(data);
      
      // トークンとユーザー情報を保存
      localStorage.setItem('token', response.token);
      setToken(response.token);
      setUser(response.user);
      
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || 'ユーザー登録に失敗しました';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  // ログアウト関数
  const logout = (): void => {
    localStorage.removeItem('token');
    setToken(null);
    setUser(null);
    setError(null);
  };

  // エラークリア関数
  const clearError = (): void => {
    setError(null);
  };

  // 認証状態の計算
  const isAuthenticated = !!user && !!token;

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated,
    loading,
    error,
    login,
    register,
    logout,
    clearError,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

// カスタムフック
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}; 