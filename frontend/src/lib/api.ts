import axios from 'axios';
import * as Types from '../types';

// Axiosクライアントの設定
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// リクエストインターセプター（認証トークンを自動的に追加）
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// レスポンスインターセプター（認証エラー時の処理）
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 認証エラーの場合、トークンを削除してログインページにリダイレクト
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// 認証関連のAPI
export const authApi = {
  // ユーザー登録
  register: async (data: Types.RegisterRequest): Promise<Types.AuthResponse> => {
    const response = await api.post('/auth/register', data);
    return response.data;
  },

  // ユーザーログイン
  login: async (data: Types.LoginRequest): Promise<Types.AuthResponse> => {
    const response = await api.post('/auth/login', data);
    return response.data;
  },

  // ユーザープロフィール取得
  getProfile: async (): Promise<Types.User> => {
    const response = await api.get('/auth/profile');
    return response.data.user;
  },
};

// ボード関連のAPI
export const boardApi = {
  // ボード一覧取得
  getBoards: async (): Promise<Types.Board[]> => {
    const response = await api.get('/boards');
    return response.data.boards;
  },

  // ボード詳細取得（カラムとタスクも含む）
  getBoardById: async (id: number): Promise<Types.Board> => {
    const response = await api.get(`/boards/${id}/columns`);
    console.log('getBoardById API response:', response.data); // デバッグログ
    console.log('getBoardById response.data.board:', response.data.board); // デバッグログ
    
    // サーバーからの形式を確認して適切に処理
    if (response.data.board) {
      return response.data.board;
    } else if (response.data.id && response.data.name) {
      // 直接返されている場合
      console.log('Board data returned directly, using response.data'); // デバッグログ
      return response.data;
    } else {
      throw new Error('Invalid response format');
    }
  },

  // ボード作成
  createBoard: async (data: Types.CreateBoardRequest): Promise<Types.Board> => {
    const response = await api.post('/boards', data);
    console.log('createBoard API response:', response.data); // デバッグログ
    console.log('createBoard response.data.board:', response.data.board); // デバッグログ
    
    // サーバーからの形式を確認して適切に処理
    if (response.data.board) {
      return response.data.board;
    } else if (response.data.id && response.data.name) {
      // 直接返されている場合
      console.log('Board data returned directly, using response.data'); // デバッグログ
      return response.data;
    } else {
      throw new Error('Invalid response format');
    }
  },

  // ボード更新
  updateBoard: async (id: number, data: Types.UpdateBoardRequest): Promise<Types.Board> => {
    const response = await api.put(`/boards/${id}`, data);
    return response.data.board;
  },

  // ボード削除
  deleteBoard: async (id: number): Promise<void> => {
    await api.delete(`/boards/${id}`);
  },
};

// タスク関連のAPI
export const taskApi = {
  // タスク作成
  createTask: async (data: Types.CreateTaskRequest): Promise<Types.Task> => {
    const response = await api.post('/tasks', data);
    console.log('createTask API response:', response.data); // デバッグログ
    console.log('createTask response.data.task:', response.data.task); // デバッグログ
    
    // サーバーからの形式を確認して適切に処理
    if (response.data.task) {
      return response.data.task;
    } else if (response.data.id && response.data.title) {
      // 直接返されている場合
      console.log('Task data returned directly, using response.data'); // デバッグログ
      return response.data;
    } else {
      throw new Error('Invalid response format');
    }
  },

  // タスク更新
  updateTask: async (id: number, data: Types.UpdateTaskRequest): Promise<Types.Task> => {
    const response = await api.put(`/tasks/${id}`, data);
    return response.data.task;
  },

  // タスク削除
  deleteTask: async (id: number): Promise<void> => {
    await api.delete(`/tasks/${id}`);
  },

  // タスク移動（ドラッグ&ドロップ）
  moveTask: async (id: number, data: Types.MoveTaskRequest): Promise<void> => {
    await api.put(`/tasks/${id}/move`, data);
  },
};

// ヘルスチェック
export const healthCheck = async (): Promise<{ service: string; status: string }> => {
  const response = await axios.get(`${API_BASE_URL}/health`);
  return response.data;
};

export default api; 