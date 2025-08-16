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

  // ボード一覧取得（カラム・タスク付き）
  getBoardsWithColumns: async (): Promise<Types.Board[]> => {
    const response = await api.get('/boards/with-columns');
    return response.data.boards;
  },

  // ボード詳細取得（カラムとタスクも含む）
  getBoardById: async (id: number): Promise<Types.Board> => {
    const response = await api.get(`/boards/${id}/columns`);
    
    // サーバーからの形式を確認して適切に処理
    if (response.data.board) {
      return response.data.board;
    } else if (response.data.id && response.data.name) {
      // 直接返されている場合
      return response.data;
    } else {
      throw new Error('Invalid response format');
    }
  },

  // ボード作成
  createBoard: async (data: Types.CreateBoardRequest): Promise<Types.Board> => {
    const response = await api.post('/boards', data);
    
    // サーバーからの形式を確認して適切に処理
    if (response.data.board) {
      return response.data.board;
    } else if (response.data.id && response.data.name) {
      // 直接返されている場合
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
    
    // サーバーからの形式を確認して適切に処理
    if (response.data.task) {
      return response.data.task;
    } else if (response.data.id && response.data.title) {
      // 直接返されている場合
      return response.data;
    } else {
      throw new Error('Invalid response format');
    }
  },

  // タスク取得
  getTask: async (id: number): Promise<Types.Task> => {
    const response = await api.get(`/tasks/${id}`);
    return response.data.task;
  },

  // タスク更新
  updateTask: async (id: number, data: Types.UpdateTaskRequest): Promise<Types.Task> => {
    const response = await api.put(`/tasks/${id}`, data);
    // サーバー実装によってはラップなしで返る場合がある
    const task = (response.data.task ?? response.data) as unknown as Types.Task;
    // サーバーのTaskResponseにはcolumn_idが含まれないため、必要最小限のフィールドを補完
    // 呼び出し側で元のtask.column_idを残す処理も併用
    return task;
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

// カレンダー関連のAPI
export const calendarApi = {
  // カレンダー設定取得
  getSettings: async (): Promise<Types.CalendarSettings> => {
    const response = await api.get('/calendar/settings');
    return response.data;
  },

  // カレンダー設定更新
  updateSettings: async (data: Types.UpdateCalendarSettingsRequest): Promise<Types.CalendarSettings> => {
    const response = await api.put('/calendar/settings', data);
    return response.data;
  },

  // カレンダーイベント取得
  getEvents: async (start: string, end: string): Promise<Types.CalendarEvent[]> => {
    const response = await api.get(`/calendar/events?start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}`);
    return response.data;
  },

  // カレンダーイベント作成
  createEvent: async (data: Omit<Types.CalendarEvent, 'id' | 'user_id' | 'created_at' | 'updated_at'>): Promise<Types.CalendarEvent> => {
    const response = await api.post('/calendar/events', data);
    return response.data;
  },

  // カレンダーイベント更新
  updateEvent: async (id: number, data: Partial<Types.CalendarEvent>): Promise<Types.CalendarEvent> => {
    const response = await api.put(`/calendar/events/${id}`, data);
    return response.data;
  },

  // カレンダーイベント削除
  deleteEvent: async (id: number): Promise<void> => {
    await api.delete(`/calendar/events/${id}`);
  },

  // タスクからカレンダーイベント作成
  createEventFromTask: async (taskId: number, start: string, end: string): Promise<void> => {
    await api.post(`/calendar/tasks/${taskId}/events`, { start, end });
  },
};

// タイマー関連のAPI
export const timerApi = {
  // タイマー開始
  startTimer: async (taskId: number, duration: number): Promise<Types.TimerSession> => {
    const response = await api.post('/timer/start', { task_id: taskId, duration });
    return response.data;
  },

  // タイマー停止
  stopTimer: async (sessionId: number): Promise<Types.TimerSession> => {
    const response = await api.put(`/timer/${sessionId}/stop`);
    return response.data;
  },

  // アクティブタイマー取得
  getActiveTimer: async (): Promise<Types.TimerSession | null> => {
    try {
      const response = await api.get('/timer/active');
      return response.data;
    } catch (error: unknown) {
      // Axiosエラーの型チェック
      const isAxiosError = error && typeof error === 'object' && 'response' in error;
      const status = isAxiosError && error.response && typeof error.response === 'object' && 'status' in error.response 
        ? (error.response as { status: number }).status 
        : null;
        
      if (status === 404) {
        return null; // アクティブなタイマーがない場合は正常（nullを返す）
      }
      
      // 404以外のエラーは再スローして呼び出し側で処理
      throw error;
    }
  },

  // タイマー履歴取得
  getTimerHistory: async (): Promise<Types.TimerSession[]> => {
    const response = await api.get('/timer/history');
    return response.data;
  },

  // タスク別タイマー履歴取得
  getTimersByTask: async (taskId: number): Promise<Types.TimerSession[]> => {
    const response = await api.get(`/timer/tasks/${taskId}`);
    return response.data;
  },
};

// 分析・統計関連のAPI
export const analyticsApi = {
  // タスク完了統計取得
  getTaskCompletionStats: async (year?: number): Promise<{ date: string; count: number }[]> => {
    const params = year ? `?year=${year}` : '';
    const response = await api.get(`/analytics/task-completion${params}`);
    return response.data;
  },
};

// ヘルスチェック
export const healthCheck = async (): Promise<{ service: string; status: string }> => {
  const response = await axios.get(`${API_BASE_URL}/health`);
  return response.data;
};

export default api; 