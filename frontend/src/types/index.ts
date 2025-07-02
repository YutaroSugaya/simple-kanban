// ユーザー関連の型定義
export interface User {
  id: string;
  email: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// タスク関連の型定義（Columnより先に定義）
export interface Task {
  id: number;
  column_id: number;
  title: string;
  description?: string;
  order: number;
  assignee_id?: string;
  due_date?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTaskRequest {
  column_id: number;
  title: string;
  description?: string;
  order: number;
  assignee_id?: string;
  due_date?: string;
}

export interface UpdateTaskRequest {
  title: string;
  description?: string;
  order: number;
  assignee_id?: string;
  due_date?: string;
}

export interface MoveTaskRequest {
  new_column_id: number;
  new_order: number;
}

// カラム関連の型定義
export interface Column {
  id: number;
  board_id: number;
  title: string;
  order: number;
  created_at: string;
  updated_at: string;
  tasks?: Task[];
}

export interface CreateColumnRequest {
  title: string;
  order: number;
}

export interface UpdateColumnRequest {
  title: string;
  order: number;
}

// ボード関連の型定義
export interface Board {
  id: number;
  name: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
  columns?: Column[];
}

export interface CreateBoardRequest {
  name: string;
}

export interface UpdateBoardRequest {
  name: string;
}

// API レスポンス関連の型定義
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface ApiError {
  error: string;
  message?: string;
}

// ドラッグ&ドロップ関連の型定義
export interface DragResult {
  draggableId: string;
  type: string;
  source: {
    droppableId: string;
    index: number;
  };
  destination: {
    droppableId: string;
    index: number;
  } | null;
}

// フォーム関連の型定義
export interface FormError {
  field: string;
  message: string;
}

// アプリケーション状態の型定義
export interface AppState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  loading: boolean;
  error: string | null;
} 